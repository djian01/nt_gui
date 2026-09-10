package testengine

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func tcpConfig(port int) Config {
	return Config{Type: "tcp", Target: "127.0.0.1", Port: port, IntervalMS: 1000, TimeoutMS: 1000}
}

func tcpListener(t *testing.T) net.Listener {
	t.Helper()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return l
}

func TestTCPProbe(t *testing.T) {
	l := tcpListener(t)
	c := tcpConfig(l.Addr().(*net.TCPAddr).Port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	p := probeTCP(ctx, c)
	if !p.Success || p.Error != "" || p.RTT < 0 {
		t.Fatalf("TCP connect: %+v", p)
	}
	l.Close()
	p = probeTCP(ctx, c)
	if p.Success || p.Error != "Conn_Refused" {
		t.Fatalf("TCP refusal: %+v", p)
	}
	cancel()
	p = probeTCP(ctx, c)
	if p.Success || !strings.Contains(p.Error, "canceled") {
		t.Fatalf("TCP cancellation: %+v", p)
	}
	expired, stop := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer stop()
	p = probeTCP(expired, c)
	if p.Success || p.Error != "Conn_Timeout" {
		t.Fatalf("TCP timeout: %+v", p)
	}
}

func TestTCPValidationAndBatch(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(c *Config) { c.Target = "" },
		func(c *Config) { c.Target = "https://localhost" },
		func(c *Config) { c.Target = "localhost:80" },
		func(c *Config) { c.Target = "invalid..host" },
		func(c *Config) { c.Port = 0 },
		func(c *Config) { c.Port = 65536 },
		func(c *Config) { c.IntervalMS = 1500 },
		func(c *Config) { c.TimeoutMS = 0 },
		func(c *Config) { c.ResolvedIP = "localhost" },
	} {
		c := tcpConfig(80)
		mutate(&c)
		if _, err := validate(c); err == nil {
			t.Fatalf("accepted invalid TCP config: %+v", c)
		}
	}
	c := tcpConfig(443)
	c.Target = "::1"
	if got, err := validate(c); err != nil || got.URL != "[::1]:443" {
		t.Fatalf("IPv6: %+v %v", got, err)
	}
	c.Target = "localhost"
	resolved, err := resolveTCP(c)
	if err != nil || net.ParseIP(resolved.ResolvedIP) == nil {
		t.Fatalf("hostname resolution: %+v %v", resolved, err)
	}
	r := newTestRunner(t, nil)
	defer r.Close()
	if _, err := r.StartTCP(c, "127.0.0.1\ninvalid/host"); err == nil {
		t.Fatal("accepted invalid second target")
	}
	if _, err := r.StartTCP(c, strings.Repeat("127.0.0.1\n", 9)); err == nil {
		t.Fatal("accepted over-capacity batch")
	}
	if len(r.sessions) != 0 {
		t.Fatal("invalid batch partially started")
	}
	created, err := r.StartTCP(c, "127.0.0.1\r\n\n::1")
	if err != nil || len(created) != 2 {
		t.Fatalf("start batch: %+v %v", created, err)
	}
	if created[0].Index != 1 || created[1].Index != 2 || created[1].Config.ResolvedIP != "::1" {
		t.Fatalf("TCP identities: %+v", created)
	}
	if _, err := r.StartTCP(c, strings.Repeat("127.0.0.1\n", 7)); err == nil {
		t.Fatal("ignored shared active capacity")
	}
	page, err := r.List("", "current-tcp", 0)
	if err != nil || len(page.Sessions) != 2 {
		t.Fatalf("TCP filter: %+v %v", page, err)
	}
	page, err = r.List("", "current-http", 0)
	if err != nil || len(page.Sessions) != 0 {
		t.Fatalf("TCP leaked into HTTP: %+v %v", page, err)
	}
	for _, s := range created {
		if _, err := r.Stop(s.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTCPRecordingLifecycle(t *testing.T) {
	l := tcpListener(t)
	updates := make(chan Session, 40)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	s, err := r.Start(tcpConfig(l.Addr().(*net.TCPAddr).Port))
	if err != nil {
		t.Fatal(err)
	}
	first := await(t, updates, func(s Session) bool { return s.Sent == 1 })
	if first.Config.Recording || !first.Last.Success || first.Config.ResolvedIP != "127.0.0.1" {
		t.Fatalf("first probe: %+v", first)
	}
	if _, err := r.store.Get(s.ID); err == nil {
		t.Fatal("unrecorded metadata reached disk")
	}
	if err := r.store.ExportCSV(s.ID, io.Discard); err == nil {
		t.Fatal("exported unrecorded test")
	}
	if err := r.Dismiss(s.ID); err == nil {
		t.Fatal("dismissed running TCP test")
	}
	if _, err := r.Record(s.ID); err != nil {
		t.Fatal(err)
	}
	await(t, updates, func(s Session) bool { return s.Sent == 2 })
	stopped, err := r.Stop(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	disk, err := r.store.Recent(s.ID)
	if err != nil || len(disk) != 1 || disk[0].Sequence != 2 {
		t.Fatalf("future recording: %+v %v", disk, err)
	}
	live, err := r.Timeline(s.ID, 0, 0)
	if err != nil || live.Count != 2 {
		t.Fatalf("live chart: %+v %v", live, err)
	}
	var exported bytes.Buffer
	if err := r.store.ExportCSV(s.ID, &exported); err != nil {
		t.Fatal(err)
	}
	imported, err := r.ImportCSV(&exported)
	if err != nil || imported.Sent != 1 || imported.Config.Type != "tcp" {
		t.Fatalf("partial import: %+v %v", imported, err)
	}
	if err := r.Dismiss(s.ID); err != nil {
		t.Fatal(err)
	}
	page, err := r.List("", "current-tcp", 0)
	if err != nil || len(page.Sessions) != 0 {
		t.Fatalf("dismissed row still current: %+v %v", page, err)
	}
	if _, err := r.store.Get(s.ID); err != nil {
		t.Fatal("dismiss deleted history", err)
	}
	replayed, err := r.Restart(stopped.ID, "")
	if err != nil || replayed.ID == s.ID || replayed.Config.Port != s.Config.Port || replayed.Config.ResolvedIP != s.Config.ResolvedIP || !replayed.Config.Recording {
		t.Fatalf("replay: %+v %v", replayed, err)
	}
	if _, err := r.Stop(replayed.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(s.ID); err == nil {
		t.Fatal("deleted TCP test still accessible")
	}
}

func TestTCPCSVRoundTripAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tcp.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	c := tcpConfig(443)
	c.Target, c.ResolvedIP, c.Recording = "server.example", "192.0.2.1", true
	c, err = validate(c)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 9, 10, 0, 0, 0, 123456, time.UTC)
	s := Session{ID: "tcp-recorded", Config: c, StartedAt: stamp, Revision: 1}
	for i := 1; i <= 4; i++ {
		p := Sample{Sequence: i, Time: stamp.Add(time.Duration(i-1) * time.Second), RTT: float64(i), Success: i != 3}
		if !p.Success {
			p.Error = "connection refused"
		}
		updateImportedSummary(&s, p)
		s.Revision++
		if err := store.Save(s, &p); err != nil {
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
	var exported bytes.Buffer
	if err := store.ExportCSV(s.ID, &exported); err != nil {
		t.Fatal(err)
	}
	imported, err := store.ImportCSV(bytes.NewReader(exported.Bytes()))
	if err != nil || imported.Sent != 4 || imported.Succeeded != 3 || imported.Config.ResolvedIP != c.ResolvedIP || !imported.Last.Time.Equal(s.Last.Time) {
		t.Fatalf("TCP roundtrip: %+v %v", imported, err)
	}
	records, err := csv.NewReader(bytes.NewReader(exported.Bytes())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var legacy bytes.Buffer
	w := csv.NewWriter(&legacy)
	for _, row := range records {
		if err := w.Write(row[:len(legacyTCPHeader)]); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	legacyImport, err := store.ImportCSV(&legacy)
	if err != nil || legacyImport.Config.TimeoutMS != 4000 || legacyImport.Sent != 4 {
		t.Fatalf("legacy TCP import: %+v %v", legacyImport, err)
	}
	records[3][5] = "80"
	var invalid bytes.Buffer
	w = csv.NewWriter(&invalid)
	w.WriteAll(records)
	before, _ := store.List("", "stopped", 0)
	if _, err := store.ImportCSV(&invalid); err == nil {
		t.Fatal("accepted changing TCP port")
	}
	after, _ := store.List("", "stopped", 0)
	if len(before.Sessions) != len(after.Sessions) {
		t.Fatal("invalid import left partial history")
	}
}

func TestTCPLegacyHostnameReplay(t *testing.T) {
	l := tcpListener(t)
	r := newTestRunner(t, nil)
	defer r.Close()
	c := tcpConfig(l.Addr().(*net.TCPAddr).Port)
	c.Target = "localhost"
	s := Session{ID: "legacy", Config: c, Sent: 1, Succeeded: 1, MinRTT: 1, AvgRTT: 1, MaxRTT: 1}
	p := Sample{Sequence: 1, Time: time.Now().UTC(), RTT: 1, Success: true}
	row := tcpCSVRow(s, s, p)[:len(legacyTCPHeader)]
	row[4] = "localhost" // legacy exported the hostname in both columns.
	var legacy bytes.Buffer
	w := csv.NewWriter(&legacy)
	w.WriteAll([][]string{legacyTCPHeader, row})
	imported, err := r.ImportCSV(&legacy)
	if err != nil || imported.Config.ResolvedIP != "" {
		t.Fatalf("legacy hostname import: %+v %v", imported, err)
	}
	// Re-exporting an unresolved legacy import must remain importable.
	var exported bytes.Buffer
	if err := r.store.ExportCSV(imported.ID, &exported); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ImportCSV(&exported); err != nil {
		t.Fatal(err)
	}
	replayed, err := r.Restart(imported.ID, "")
	if err != nil || net.ParseIP(replayed.Config.ResolvedIP) == nil || replayed.Config.Target != "localhost" {
		t.Fatalf("legacy replay did not resolve hostname: %+v %v", replayed, err)
	}
	if _, err := r.Stop(replayed.ID); err != nil {
		t.Fatal(err)
	}
}

func TestTCPBatchStorageFailureRollsBack(t *testing.T) {
	r := newTestRunner(t, nil)
	defer r.Close()
	// Fail persistence of the second worker after the first has been prepared.
	_, err := r.store.db.Exec(`CREATE TRIGGER fail_second_tcp BEFORE INSERT ON tests
	WHEN (SELECT count(*) FROM tests)=1 BEGIN SELECT RAISE(FAIL, 'fixture failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	c := tcpConfig(1)
	c.Recording = true
	if _, err := r.StartTCP(c, "127.0.0.1\n::1"); err == nil {
		t.Fatal("expected storage failure")
	}
	r.mu.Lock()
	remaining := len(r.sessions)
	r.mu.Unlock()
	if remaining != 0 {
		t.Fatal("failed batch left workers")
	}
	for _, filter := range []string{"current-tcp", "all"} {
		page, err := r.List("", filter, 0)
		if err != nil || len(page.Sessions) != 0 {
			t.Fatalf("failed batch left %s entries: %+v %v", filter, page, err)
		}
	}
}
