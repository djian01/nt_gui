package testengine

import (
	"bytes"
	"context"
	"encoding/csv"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func icmpConfig() Config {
	return Config{Type: "icmp", Target: "127.0.0.1", ResolvedIP: "127.0.0.1", PayloadSize: 32, IntervalMS: 1000, TimeoutMS: 1000}
}
func TestICMPRecordingLifecycle(t *testing.T) {
	updates := make(chan Session, 40)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	s, err := r.Start(icmpConfig())
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
		t.Fatal("dismissed running ICMP test")
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
	if err != nil || imported.Sent != 1 || imported.Config.Type != "icmp" {
		t.Fatalf("partial import: %+v %v", imported, err)
	}
	if err := r.Dismiss(s.ID); err != nil {
		t.Fatal(err)
	}
	page, err := r.List("", "current-icmp", 0)
	if err != nil || len(page.Sessions) != 0 {
		t.Fatalf("dismissed row still current: %+v %v", page, err)
	}
	if _, err := r.store.Get(s.ID); err != nil {
		t.Fatal("dismiss deleted history", err)
	}
	replayed, err := r.Restart(stopped.ID, "")
	if err != nil || replayed.ID == s.ID || replayed.Config.PayloadSize != s.Config.PayloadSize || replayed.Config.ResolvedIP != s.Config.ResolvedIP || !replayed.Config.Recording {
		t.Fatalf("replay: %+v %v", replayed, err)
	}
	if _, err := r.Stop(replayed.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(s.ID); err == nil {
		t.Fatal("deleted ICMP test still accessible")
	}
}

func TestICMPCSVRoundTripAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "icmp.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	c := icmpConfig()
	c.Target, c.ResolvedIP, c.Recording = "server.example", "192.0.2.1", true
	c.DF = true
	c, err = validate(c)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 9, 10, 0, 0, 0, 123456, time.UTC)
	s := Session{ID: "icmp-recorded", Config: c, StartedAt: stamp, Revision: 1}
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
	if err != nil || imported.Sent != 4 || imported.Succeeded != 3 || imported.Config.ResolvedIP != c.ResolvedIP || !imported.Config.DF || !imported.Last.Time.Equal(s.Last.Time) {
		t.Fatalf("ICMP roundtrip: %+v %v", imported, err)
	}
	records, err := csv.NewReader(bytes.NewReader(exported.Bytes())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var legacy bytes.Buffer
	w := csv.NewWriter(&legacy)
	for _, row := range records {
		if err := w.Write(row[:len(legacyICMPHeader)]); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	legacyImport, err := store.ImportCSV(&legacy)
	if err != nil || legacyImport.Config.TimeoutMS != 4000 || legacyImport.Config.DF || legacyImport.Sent != 4 {
		t.Fatalf("legacy ICMP import: %+v %v", legacyImport, err)
	}
	records[3][5] = "64"
	var invalid bytes.Buffer
	w = csv.NewWriter(&invalid)
	w.WriteAll(records)
	before, _ := store.List("", "stopped", 0)
	if _, err := store.ImportCSV(&invalid); err == nil {
		t.Fatal("accepted changing ICMP payload")
	}
	after, _ := store.List("", "stopped", 0)
	if len(before.Sessions) != len(after.Sessions) {
		t.Fatal("invalid import left partial history")
	}
}

func TestICMPLegacyHostnameReplay(t *testing.T) {
	r := newTestRunner(t, nil)
	defer r.Close()
	c := icmpConfig()
	c.Target = "localhost"
	s := Session{ID: "legacy", Config: c, Sent: 1, Succeeded: 1, MinRTT: 1, AvgRTT: 1, MaxRTT: 1}
	p := Sample{Sequence: 1, Time: time.Now().UTC(), RTT: 1, Success: true}
	row := icmpCSVRow(s, s, p)[:len(legacyICMPHeader)]
	row[4] = "localhost" // legacy exported the hostname in both columns.
	var legacy bytes.Buffer
	w := csv.NewWriter(&legacy)
	w.WriteAll([][]string{legacyICMPHeader, row})
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

func TestICMPUnprivilegedProbe(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("regular-user integration check is not applicable as root")
	}
	for _, df := range []bool{false, true} {
		c := icmpConfig()
		c.DF = df
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		sample := probeICMP(ctx, c)
		cancel()
		if !sample.Success || sample.RTT <= 0 {
			t.Fatalf("unprivileged DF=%v: %+v", df, sample)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if p := probeICMP(ctx, icmpConfig()); p.Success || p.Error == "" {
		t.Fatalf("canceled probe: %+v", p)
	}
}

func TestICMPValidationBatchAndFilters(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(c *Config) { c.Target = "" }, func(c *Config) { c.Target = "https://localhost" },
		func(c *Config) { c.Target = "bad/host" }, func(c *Config) { c.Target = "::1" },
		func(c *Config) { c.PayloadSize = 31 }, func(c *Config) { c.PayloadSize = 65508 },
		func(c *Config) { c.TimeoutMS = 0 }, func(c *Config) { c.IntervalMS = 1500 },
		func(c *Config) { c.ResolvedIP = "::1" },
	} {
		c := icmpConfig()
		mutate(&c)
		if _, err := validate(c); err == nil {
			t.Fatalf("accepted invalid config: %+v", c)
		}
	}
	r := newTestRunner(t, nil)
	defer r.Close()
	for _, targets := range []string{"", "127.0.0.1\ninvalid/host", strings.Repeat("127.0.0.1\n", 9)} {
		if _, err := r.StartICMP(icmpConfig(), targets); err == nil {
			t.Fatalf("accepted invalid batch %q", targets)
		}
	}
	if len(r.sessions) != 0 {
		t.Fatal("invalid batch started probes")
	}
	created, err := r.StartICMP(icmpConfig(), "127.0.0.1\r\n\nlocalhost")
	if err != nil || len(created) != 2 {
		t.Fatalf("batch: %+v %v", created, err)
	}
	for i, s := range created {
		if s.Index != i+1 || net.ParseIP(s.Config.ResolvedIP).To4() == nil {
			t.Fatalf("identity: %+v", s)
		}
	}
	for _, filter := range []string{"current-icmp", "current-http", "current-tcp", "current-dns"} {
		page, err := r.List("", filter, 0)
		want := 0
		if filter == "current-icmp" {
			want = 2
		}
		if err != nil || len(page.Sessions) != want {
			t.Fatalf("%s: %+v %v", filter, page, err)
		}
	}
	if _, err := r.StartICMP(icmpConfig(), strings.Repeat("127.0.0.1\n", 7)); err == nil {
		t.Fatal("ignored shared active limit")
	}
	for _, s := range created {
		if _, err := r.Stop(s.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func TestICMPReplyCorrelation(t *testing.T) {
	payload := bytes.Repeat([]byte{7}, 32)
	reply := icmp.Message{Type: ipv4.ICMPTypeEchoReply, Body: &icmp.Echo{ID: 17, Seq: 1, Data: payload}}
	wire, err := reply.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}
	peer := &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}
	if !matchesICMPEcho(wire, peer, "127.0.0.1", payload) {
		t.Fatal("valid reply rejected")
	}
	if matchesICMPEcho(wire, peer, "192.0.2.1", payload) || matchesICMPEcho(wire, peer, "127.0.0.1", bytes.Repeat([]byte{8}, 32)) || matchesICMPEcho([]byte{1}, peer, "127.0.0.1", payload) {
		t.Fatal("unrelated reply accepted")
	}
}

func TestICMPCommandArgumentsAndErrors(t *testing.T) {
	c := icmpConfig()
	c.DF = true
	c.TimeoutMS = 4000
	for platform, want := range map[string][]string{
		"darwin":  {"-n", "-c", "1", "-W", "4000", "-s", "32", "-D", "127.0.0.1"},
		"linux":   {"-n", "-c", "1", "-W", "4", "-s", "32", "-M", "do", "127.0.0.1"},
		"windows": {"-4", "-n", "1", "-w", "4000", "-l", "32", "-f", "127.0.0.1"},
	} {
		args, err := icmpCommandArgs(c, platform)
		if err != nil || !reflect.DeepEqual(args, want) {
			t.Fatalf("%s: %v %v", platform, args, err)
		}
	}
	for output, want := range map[string]string{
		"ping: sendto: Message too long": "MTU Exceed, DF set", "Packet needs to be fragmented but DF set.": "MTU Exceed, DF set",
		"Destination host unreachable": "NoRoute", "Request timed out.": "Timeout", "100.0% packet loss": "Timeout", "Time to live exceeded": "TTL_Exceeded",
	} {
		ok, rtt, info := parseICMPOutput(output)
		if ok || rtt != 0 || info != want {
			t.Fatalf("%q: %v %v %q", output, ok, rtt, info)
		}
	}
	for _, output := range []string{"64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.123 ms", "Reply from 127.0.0.1: bytes=32 time<1ms TTL=128"} {
		ok, rtt, info := parseICMPOutput(output)
		if !ok || rtt <= 0 || info != "" {
			t.Fatalf("reply: %v %v %q", ok, rtt, info)
		}
	}
	var output pingOutput
	n, err := output.Write(bytes.Repeat([]byte{'x'}, 20000))
	if n != 20000 || err != nil || output.Len() != 8192 {
		t.Fatal("unbounded command output")
	}
}
