package testengine

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store owns a fresh Wails database. It never opens or migrates legacy data.
type Store struct{ db *sql.DB }

type Page struct {
	Sessions []Session `json:"sessions"`
	Next     int64     `json:"next"`
	Active   int       `json:"active"`
}

type Timeline struct {
	Samples    []Sample `json:"samples"`
	Count      int      `json:"count"`
	Succeeded  int      `json:"succeeded"`
	Average    float64  `json:"average"`
	Maximum    float64  `json:"maximum"`
	From       int64    `json:"from"`
	To         int64    `json:"to"`
	Revision   int      `json:"revision"`
	Aggregated bool     `json:"aggregated"`
}

func DefaultStorePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nt-wails", "results.db"), nil
}

func OpenStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	f.Close()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err = s.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initialize() error {
	for _, pragma := range []string{"PRAGMA busy_timeout=1000", "PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA foreign_keys=ON"} {
		if _, err := s.db.Exec(pragma); err != nil {
			return err
		}
	}
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version > 1 {
		return errors.New("This results database requires a newer application")
	}
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS tests (id TEXT NOT NULL UNIQUE, url TEXT NOT NULL, running INTEGER NOT NULL, data TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS tests_running ON tests(running);
	CREATE TEMP TABLE current_tests (id TEXT PRIMARY KEY);
	CREATE TABLE IF NOT EXISTS samples (
	 test_id TEXT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
	 seq INTEGER NOT NULL, time INTEGER NOT NULL, data TEXT NOT NULL,
	 PRIMARY KEY(test_id, seq));
	CREATE INDEX IF NOT EXISTS samples_time ON samples(test_id, time, seq);
	CREATE TABLE IF NOT EXISTS buckets (
	 test_id TEXT NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
	 level INTEGER NOT NULL, bucket INTEGER NOT NULL,
	 count INTEGER NOT NULL, good INTEGER NOT NULL, total REAL NOT NULL,
	 min REAL NOT NULL, max REAL NOT NULL,
	 first TEXT NOT NULL, last TEXT NOT NULL, low TEXT NOT NULL, high TEXT NOT NULL, failure TEXT NOT NULL,
	 PRIMARY KEY(test_id, level, bucket));
	PRAGMA user_version=1;`)
	if err != nil {
		return err
	}
	// A previous process cannot still own this file: the desktop uses a single instance.
	rows, err := s.db.Query("SELECT data FROM tests WHERE running=1")
	if err != nil {
		return err
	}
	var interrupted []Session
	for rows.Next() {
		var data []byte
		var test Session
		if err = rows.Scan(&data); err != nil {
			break
		}
		if err = json.Unmarshal(data, &test); err != nil {
			break
		}
		test.Running = false
		test.EndReason = "interrupted"
		test.Revision++
		end := test.StartedAt
		if test.Last != nil {
			end = test.Last.Time
		}
		test.EndedAt = &end
		interrupted = append(interrupted, test)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return err
	}
	for _, test := range interrupted {
		if err := s.Save(test, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Save(test Session, sample *Sample) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := saveTest(tx, test, sample); err != nil {
		return err
	}
	return tx.Commit()
}

func saveTest(tx *sql.Tx, test Session, sample *Sample) error {
	// Never serialize transport state or proxy passwords.
	test.Config.Proxy.Password = ""
	data, err := json.Marshal(test)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO tests(id,url,running,data) VALUES(?,?,?,?)
	 ON CONFLICT(id) DO UPDATE SET running=excluded.running,data=excluded.data`, test.ID, test.Config.URL, test.Running, string(data))
	if err != nil {
		return err
	}
	// Session membership is connection-local and survives stopping, but not app exit.
	if test.Running && test.Revision == 1 {
		if _, err := tx.Exec("INSERT OR IGNORE INTO current_tests(id) VALUES(?)", test.ID); err != nil {
			return err
		}
	}
	if sample != nil {
		if err := saveSample(tx, test.ID, sample); err != nil {
			return err
		}
	}
	return nil
}

func saveSample(tx *sql.Tx, id string, sample *Sample) error {
	point, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO samples(test_id,seq,time,data) VALUES(?,?,?,?)", id, sample.Sequence, sample.Time.UnixMilli(), string(point))
	if err != nil {
		return err
	}
	good, total, low, high, failure := 0, 0.0, "", "", ""
	if sample.Success {
		good = 1
		total = sample.RTT
		low = string(point)
		high = string(point)
	} else {
		failure = string(point)
	}
	// A radix-4 hierarchy has approximately one bucket per three raw samples.
	// Updating fixed levels avoids rebuilding the overview as a test grows.
	statement, err := tx.Prepare(`INSERT INTO buckets VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)
		 ON CONFLICT(test_id,level,bucket) DO UPDATE SET
		 count=buckets.count+1, good=buckets.good+excluded.good,total=buckets.total+excluded.total,
		 low=CASE WHEN excluded.good=1 AND (buckets.good=0 OR excluded.min<buckets.min) THEN excluded.low ELSE buckets.low END,
		 high=CASE WHEN excluded.good=1 AND (buckets.good=0 OR excluded.max>buckets.max) THEN excluded.high ELSE buckets.high END,
		 min=CASE WHEN excluded.good=1 AND (buckets.good=0 OR excluded.min<buckets.min) THEN excluded.min ELSE buckets.min END,
		 max=CASE WHEN excluded.good=1 AND (buckets.good=0 OR excluded.max>buckets.max) THEN excluded.max ELSE buckets.max END,
		 last=excluded.last, failure=CASE WHEN buckets.failure='' THEN excluded.failure ELSE buckets.failure END`)
	if err != nil {
		return err
	}
	defer statement.Close()
	for level, size := 1, int64(4); level <= 20; level, size = level+1, size*4 {
		_, err = statement.Exec(id, level, int64(sample.Sequence-1)/size, 1, good, total, total, total,
			string(point), string(point), low, high, failure)
		if err != nil {
			return err
		}
	}
	return nil
}

func readSession(row *sql.Row) (Session, error) {
	var data []byte
	var test Session
	err := row.Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return test, errors.New("Test not found")
	}
	if err == nil {
		err = json.Unmarshal(data, &test)
	}
	return test, err
}

func (s *Store) Get(id string) (Session, error) {
	return readSession(s.db.QueryRow("SELECT data FROM tests WHERE id=?", id))
}

func (s *Store) List(search, filter string, before int64) (Page, error) {
	p := Page{Sessions: []Session{}}
	where, args := "WHERE 1=1", []any{}
	// Apply protocol selection before the History cursor and page limit.
	switch filter {
	case "stopped-http", "stopped-dns", "stopped-tcp", "stopped-icmp":
		where += " AND coalesce(nullif(json_extract(data,'$.config.type'),''),'http')=?"
		args = append(args, strings.TrimPrefix(filter, "stopped-"))
		filter = "stopped"
	}
	// Older builds wrote unrecorded summaries. Keep them out of History
	// without deleting existing data; missing type is legacy recorded HTTP.
	if filter == "stopped" || filter == "all" || filter == "" {
		where += " AND (coalesce(json_extract(data,'$.config.type'),'')='' OR json_extract(data,'$.config.recording')=1)"
	}
	if filter == "current-dns" || filter == "current-http" || filter == "current-tcp" || filter == "current-icmp" {
		if filter == "current-dns" || filter == "current-tcp" || filter == "current-icmp" {
			where += " AND json_extract(data,'$.config.type')=?"
			args = append(args, strings.TrimPrefix(filter, "current-"))
		} else {
			where += " AND coalesce(json_extract(data,'$.config.type'),'http')='http'"
		}
		filter = "current"
	}
	if filter == "current" {
		where += " AND id IN (SELECT id FROM current_tests)"
	} else if filter == "running" {
		where += " AND running=1"
	} else if filter == "stopped" {
		where += " AND running=0"
	}
	if search != "" {
		where += " AND instr(lower(url),?)>0"
		args = append(args, strings.ToLower(search))
	}
	if before > 0 {
		where += " AND rowid<?"
		args = append(args, before)
	}
	rows, err := s.db.Query("SELECT rowid,data FROM tests "+where+" ORDER BY rowid DESC LIMIT 51", args...)
	if err != nil {
		return p, err
	}
	defer rows.Close()
	var last int64
	for rows.Next() {
		var rowid int64
		var data []byte
		var test Session
		if err := rows.Scan(&rowid, &data); err != nil {
			return p, err
		}
		if len(p.Sessions) == 50 {
			p.Next = last
			break
		}
		if err := json.Unmarshal(data, &test); err != nil {
			return p, err
		}
		p.Sessions = append(p.Sessions, test)
		last = rowid
	}
	return p, rows.Err()
}

func (s *Store) Recent(id string) ([]Sample, error) {
	rows, err := s.db.Query("SELECT data FROM samples WHERE test_id=? ORDER BY seq DESC LIMIT 6", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Sample{}
	for rows.Next() {
		var data []byte
		var p Sample
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Sequence < result[j].Sequence })
	return result, rows.Err()
}

// discardUnstarted is only used for aborted batches while the runner lock
// prevents probe delivery. Unlike user deletion it may remove a running marker.
func (s *Store) discardUnstarted(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM tests WHERE id=?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM current_tests WHERE id=?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Remove(id string) error {
	result, err := s.db.Exec("DELETE FROM tests WHERE id=? AND running=0", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return errors.New("Stop the test before deleting its saved history")
	}
	return err
}

func (s *Store) Timeline(id string, from, to int64) (Timeline, error) {
	out := Timeline{Samples: []Sample{}}
	tx, err := s.db.Begin()
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	test, err := readSession(tx.QueryRow("SELECT data FROM tests WHERE id=?", id))
	if err != nil {
		return out, err
	}
	out.Revision = test.Revision
	if from == 0 {
		from = test.StartedAt.UnixMilli()
	}
	if to == 0 {
		to = time.Now().UnixMilli()
		if test.EndedAt != nil {
			to = test.EndedAt.UnixMilli()
		}
	}
	if to < from {
		return out, errors.New("Timeline end must follow its start")
	}
	out.From = from
	out.To = to
	var first, last sql.NullInt64
	err = tx.QueryRow("SELECT seq FROM samples WHERE test_id=? AND time>=? AND time<=? ORDER BY time,seq LIMIT 1", id, from, to).Scan(&first)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	err = tx.QueryRow("SELECT seq FROM samples WHERE test_id=? AND time>=? AND time<=? ORDER BY time DESC,seq DESC LIMIT 1", id, from, to).Scan(&last)
	if err != nil {
		return out, err
	}
	count := last.Int64 - first.Int64 + 1
	levelLimit, sizeLimit := 0, int64(1)
	for count/sizeLimit > 100 && levelLimit < 20 {
		levelLimit++
		sizeLimit *= 4
	}
	// Cover the exact sequence range with complete aligned buckets. Small edge
	// buckets retain exact range statistics without scanning excluded samples.
	seen := map[int]bool{}
	add := func(data string) error {
		if data == "" {
			return nil
		}
		var p Sample
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return err
		}
		if !seen[p.Sequence] {
			seen[p.Sequence] = true
			out.Samples = append(out.Samples, p)
		}
		return nil
	}
	var total float64
	for pos := first.Int64 - 1; pos < last.Int64; {
		level, size := levelLimit, sizeLimit
		for level > 0 && (pos%size != 0 || pos+size > last.Int64) {
			level--
			size /= 4
		}
		if level == 0 {
			var data string
			var p Sample
			if err := tx.QueryRow("SELECT data FROM samples WHERE test_id=? AND seq=?", id, pos+1).Scan(&data); err != nil {
				return out, err
			}
			if err := json.Unmarshal([]byte(data), &p); err != nil {
				return out, err
			}
			out.Count++
			if p.Success {
				out.Succeeded++
				total += p.RTT
				out.Maximum = max(out.Maximum, p.RTT)
			}
			if err := add(data); err != nil {
				return out, err
			}
		} else {
			var n, good int
			var sum, maximum float64
			var a, b, lo, hi, fail string
			err := tx.QueryRow("SELECT count,good,total,max,first,last,low,high,failure FROM buckets WHERE test_id=? AND level=? AND bucket=?", id, level, pos/size).Scan(&n, &good, &sum, &maximum, &a, &b, &lo, &hi, &fail)
			if err != nil {
				return out, err
			}
			out.Count += n
			out.Succeeded += good
			total += sum
			out.Maximum = max(out.Maximum, maximum)
			out.Aggregated = true
			for _, data := range []string{a, b, lo, hi, fail} {
				if err := add(data); err != nil {
					return out, err
				}
			}
		}
		pos += size
	}
	if out.Succeeded > 0 {
		out.Average = total / float64(out.Succeeded)
	}
	sort.Slice(out.Samples, func(i, j int) bool { return out.Samples[i].Sequence < out.Samples[j].Sequence })
	return out, tx.Commit()
}

// ExportCSV pages raw samples to a fixed sequence watermark, independent of zoom.
func (s *Store) ExportCSV(id string, w io.Writer) error {
	test, err := s.Get(id)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(w)
	header := wailsCSVFullHeader
	if !recordingEnabled(test.Config) {
		return errors.New("This test has no recording; enable recording while it is running")
	}
	if test.Config.Type == "dns" {
		header = dnsCSVHeader
	} else if test.Config.Type == "icmp" {
		header = icmpCSVHeader
	} else if test.Config.Type == "tcp" {
		header = tcpCSVHeader
	}
	if test.Sent == 0 {
		return errors.New("No probes have been recorded yet")
	}
	if err = writer.Write(header); err != nil {
		return err
	}
	var summary Session
	for after := 0; after < test.Sent; {
		rows, err := s.db.Query("SELECT data FROM samples WHERE test_id=? AND seq>? AND seq<=? ORDER BY seq LIMIT 256", id, after, test.Sent)
		if err != nil {
			return err
		}
		var batch []Sample
		for rows.Next() {
			var data []byte
			var p Sample
			if err = rows.Scan(&data); err != nil {
				break
			}
			if err = json.Unmarshal(data, &p); err != nil {
				break
			}
			batch = append(batch, p)
		}
		if err == nil {
			err = rows.Err()
		}
		rows.Close()
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			if after == 0 {
				return errors.New("No probes have been recorded yet")
			}
			return errors.New("Test was deleted during export")
		}
		for _, p := range batch {
			if test.Config.Type == "dns" {
				updateImportedSummary(&summary, p)
				err = writer.Write(dnsCSVRow(test, summary, p))
			} else if test.Config.Type == "icmp" {
				updateImportedSummary(&summary, p)
				err = writer.Write(icmpCSVRow(test, summary, p))
			} else if test.Config.Type == "tcp" {
				updateImportedSummary(&summary, p)
				err = writer.Write(tcpCSVRow(test, summary, p))
			} else {
				err = writer.Write([]string{test.ID, test.Config.URL, test.Config.Method, strconv.Itoa(test.Config.IntervalMS), strconv.Itoa(test.Config.TimeoutMS), strconv.Itoa(p.Sequence), p.Time.UTC().Format(time.RFC3339Nano), strconv.FormatFloat(p.RTT, 'f', 3, 64), strconv.Itoa(p.StatusCode), strconv.FormatBool(p.Success), p.Error, p.ResponsePhase, strings.Join(test.Config.AcceptedStatuses, ","), test.Config.Proxy.URL, test.Config.Proxy.Username, strconv.FormatBool(test.PasswordRequired), strconv.FormatBool(test.Config.FollowRedirects)})
			}
			if err != nil {
				return err
			}
			after = p.Sequence
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("Export failed: %w", err)
	}
	return nil
}
