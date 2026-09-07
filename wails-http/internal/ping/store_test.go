package ping

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSavedHistoryTimelineAndExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	test := Session{ID: "history", Config: config("https://example.com/health"), StartedAt: start, Running: true, Revision: 1}
	test.Config.Proxy.Password = "never-on-disk"
	const count = 4200
	raw := make([]Sample, count)
	for i := range raw {
		p := Sample{Sequence: i + 1, Time: start.Add(time.Duration(i) * time.Second), RTT: float64((i*31)%997) + 0.5, Success: i%23 != 0, StatusCode: 200}
		if !p.Success {
			p.StatusCode = 503
			p.Error = "Service unavailable"
		}
		raw[i] = p
		test.Sent++
		test.Last = &p
		test.Revision++
		if p.Success {
			test.Succeeded++
			test.AvgRTT += (p.RTT - test.AvgRTT) / float64(test.Succeeded)
			test.MaxRTT = max(test.MaxRTT, p.RTT)
			if test.Succeeded == 1 {
				test.MinRTT = p.RTT
			} else {
				test.MinRTT = min(test.MinRTT, p.RTT)
			}
		}
		if err := store.Save(test, &p); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	restored, err := store.Get(test.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Running || restored.EndReason != "interrupted" || restored.Sent != count || restored.Config.Proxy.Password != "" {
		t.Fatalf("incorrect recovery: %+v", restored)
	}
	if restored.EndedAt == nil || !restored.EndedAt.Equal(raw[count-1].Time) {
		t.Fatal("interruption boundary differs from last saved probe")
	}
	for _, bounds := range [][2]int{{0, count - 1}, {1, 4198}, {59, 3131}, {255, 257}, {333, 333}, {17, 29}} {
		from, to := bounds[0], bounds[1]
		timeline, err := store.Timeline(test.ID, raw[from].Time.UnixMilli(), raw[to].Time.UnixMilli())
		if err != nil {
			t.Fatal(err)
		}
		var good int
		var sum, maximum float64
		for _, p := range raw[from : to+1] {
			if p.Success {
				good++
				sum += p.RTT
				maximum = max(maximum, p.RTT)
			}
		}
		if timeline.Count != to-from+1 || timeline.Succeeded != good || timeline.Maximum != maximum {
			t.Fatalf("range %v incorrect: %+v", bounds, timeline)
		}
		if good > 0 && math.Abs(timeline.Average-sum/float64(good)) > 0.000001 {
			t.Fatal("overview average differs from raw probes")
		}
		if len(timeline.Samples) > 800 {
			t.Fatalf("unbounded overview: %d", len(timeline.Samples))
		}
		previous := 0
		hasPeak := false
		for _, p := range timeline.Samples {
			if p.Sequence <= previous {
				t.Fatal("out of order or duplicated point")
			}
			previous = p.Sequence
			if p.Success && p.RTT == maximum {
				hasPeak = true
			}
		}
		if good > 0 && !hasPeak {
			t.Fatal("overview lost the latency peak")
		}
	}
	var buffer bytes.Buffer
	if err := store.ExportCSV(test.ID, &buffer); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buffer.String(), "never-on-disk") {
		t.Fatal("export contains a proxy secret")
	}
	records, err := csv.NewReader(&buffer).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != count+1 || records[1][5] != "1" || records[count][5] != fmt.Sprint(count) {
		t.Fatal("export truncated or reordered saved history")
	}
	if err := store.Remove(test.ID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"samples", "buckets"} {
		var n int
		if err := store.db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("delete left %s behind", table)
		}
	}
}

func TestHistoryPaginationAndCleanShutdown(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for i := 0; i < 125; i++ {
		s := Session{ID: fmt.Sprint(i), Config: config(fmt.Sprintf("https://example.com/%d", i)), StartedAt: time.Now()}
		if err := store.Save(s, nil); err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	var before int64
	for {
		page, err := store.List("example", "stopped", before)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Sessions) > 50 {
			t.Fatal("page unbounded")
		}
		for _, s := range page.Sessions {
			if seen[s.ID] {
				t.Fatal("duplicate across pages")
			}
			seen[s.ID] = true
		}
		if page.Next == 0 {
			break
		}
		before = page.Next
	}
	if len(seen) != 125 {
		t.Fatal("history lost tests beyond old 24-session limit")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer server.Close()
	r := New(store, nil)
	s, err := r.Start(config(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = store.Get(s.ID)
	if err != nil || s.Running || s.EndReason != "stopped" {
		t.Fatalf("clean shutdown not saved: %+v %v", s, err)
	}
}

func TestSavingFailureStopsProbeWithoutPartialCommit(t *testing.T) {
	r := newTestRunner(t, nil)
	defer r.Close()
	_, err := r.store.db.Exec("CREATE TRIGGER reject_probe BEFORE INSERT ON samples BEGIN SELECT RAISE(ABORT,'disk failure fixture'); END")
	if err != nil {
		t.Fatal(err)
	}
	updates := make(chan Session, 10)
	r.onUpdate = func(s Session) { updates <- s }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer server.Close()
	s, err := r.Start(config(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	failed := await(t, updates, func(s Session) bool { return s.EndReason == "storage_error" })
	if failed.Running || failed.Sent != 0 || failed.SaveError == "" {
		t.Fatalf("storage failure hidden: %+v", failed)
	}
	saved, err := r.store.Get(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Sent != 0 || saved.Running {
		t.Fatal("partial sample/summary transaction committed")
	}
	var n int
	r.store.db.QueryRow("SELECT count(*) FROM samples").Scan(&n)
	if n != 0 {
		t.Fatal("failed transaction left a sample")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failure") }

func TestExportErrorsAndSchemaVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results.db")
	s, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(Session{ID: "empty", Config: config("https://example.com")}, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.ExportCSV("empty", failingWriter{}); err == nil {
		t.Fatal("export swallowed write failure")
	}
	if _, err := s.db.Exec("PRAGMA user_version=99"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if newer, err := OpenStore(path); err == nil {
		newer.Close()
		t.Fatal("opened an unsupported schema")
	}
}
