package testengine

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/csv"
	"io"
	"net"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func dnsConfig() Config {
	return Config{Type: "dns", Resolver: "127.0.0.1", Query: "invalid..domain", Protocol: "udp", IntervalMS: 1000, TimeoutMS: 1000}
}

func TestDNSValidationAndBatch(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(c *Config) { c.Resolver = "resolver.example" },
		func(c *Config) { c.Resolver = "127.0.0.1:53" },
		func(c *Config) { c.Query = "" },
		func(c *Config) { c.Protocol = "https" },
		func(c *Config) { c.IntervalMS = 1500 },
		func(c *Config) { c.TimeoutMS = 0 },
	} {
		c := dnsConfig()
		mutate(&c)
		if _, err := validate(c); err == nil {
			t.Fatalf("accepted invalid DNS config: %+v", c)
		}
	}
	c := dnsConfig()
	c.Resolver, c.Protocol = "::1", ""
	if got, err := validate(c); err != nil || got.Protocol != "udp" {
		t.Fatalf("IPv6/default protocol: %+v %v", got, err)
	}
	r := newTestRunner(t, nil)
	defer r.Close()
	if _, err := r.StartDNS(dnsConfig(), "127.0.0.1\ninvalid"); err == nil {
		t.Fatal("accepted invalid second resolver")
	}
	if len(r.sessions) != 0 {
		t.Fatal("partially started invalid batch")
	}
	if _, err := r.StartDNS(dnsConfig(), strings.Repeat("127.0.0.1\n", 9)); err == nil {
		t.Fatal("accepted over-capacity batch")
	}
	created, err := r.StartDNS(dnsConfig(), "127.0.0.1\r\n\n::1")
	if err != nil || len(created) != 2 {
		t.Fatalf("batch start: %+v %v", created, err)
	}
	if created[0].Index != 1 || created[1].Index != 2 {
		t.Fatal("DNS row indices are not independent and consecutive")
	}
	for _, s := range created {
		if _, err := r.Stop(s.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDNSRecordingLifecycle(t *testing.T) {
	updates := make(chan Session, 40)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	s, err := r.Start(dnsConfig())
	if err != nil {
		t.Fatal(err)
	}
	first := await(t, updates, func(s Session) bool { return s.Sent == 1 })
	if first.Config.Recording || first.Last.Success {
		t.Fatalf("unexpected unrecorded invalid-domain probe: %+v", first)
	}
	live, err := r.Timeline(s.ID, 0, 0)
	if err != nil || live.Count != 1 {
		t.Fatalf("live unrecorded timeline: %+v %v", live, err)
	}
	disk, err := r.store.Recent(s.ID)
	if err != nil || len(disk) != 0 {
		t.Fatalf("unrecorded probes reached disk: %+v %v", disk, err)
	}
	if err := r.store.ExportCSV(s.ID, io.Discard); err == nil {
		t.Fatal("exported unrecorded test")
	}
	if err := r.Dismiss(s.ID); err == nil {
		t.Fatal("dismissed running DNS row")
	}
	if recorded, err := r.Record(s.ID); err != nil || !recorded.Config.Recording {
		t.Fatalf("enable recording: %+v %v", recorded, err)
	}
	await(t, updates, func(s Session) bool { return s.Sent == 2 })
	stopped, err := r.Stop(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	disk, err = r.store.Recent(s.ID)
	if err != nil || len(disk) != 1 || disk[0].Sequence != 2 {
		t.Fatalf("recording should include only future probes: %+v %v", disk, err)
	}
	live, err = r.Timeline(s.ID, 0, 0)
	if err != nil || live.Count != 2 {
		t.Fatalf("earlier chart probes lost: %+v %v", live, err)
	}
	var exported bytes.Buffer
	if err := r.store.ExportCSV(s.ID, &exported); err != nil {
		t.Fatal(err)
	}
	imported, err := r.ImportCSV(bytes.NewReader(exported.Bytes()))
	if err != nil || imported.Sent != 1 || imported.Config.Type != "dns" {
		t.Fatalf("partial recording import: %+v %v", imported, err)
	}
	if err := r.Dismiss(s.ID); err != nil {
		t.Fatal(err)
	}
	page, err := r.List("", "current-dns", 0)
	if err != nil || len(page.Sessions) != 0 {
		t.Fatalf("dismissed row still current: %+v %v", page, err)
	}
	if _, err := r.store.Get(s.ID); err != nil {
		t.Fatal("dismiss deleted history", err)
	}
	replayed, err := r.Restart(stopped.ID, "")
	if err != nil || replayed.ID == stopped.ID || !replayed.Config.Recording || replayed.Config.Protocol != stopped.Config.Protocol {
		t.Fatalf("DNS replay lost settings: %+v %v", replayed, err)
	}
	if _, err := r.Stop(replayed.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(s.ID); err == nil {
		t.Fatal("deleted DNS test still accessible")
	}
}

func TestDNSCSVLegacyRoundTripAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dns.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	c, err := validate(dnsConfig())
	if err != nil {
		t.Fatal(err)
	}
	c.Recording = true
	stamp := time.Date(2026, 9, 9, 0, 0, 0, 123456, time.UTC)
	s := Session{ID: "dns-recorded", Config: c, StartedAt: stamp, Revision: 1}
	for i := 1; i <= 4; i++ {
		p := Sample{Sequence: i, Time: stamp.Add(time.Duration(i-1) * time.Second), RTT: float64(i), Success: i != 3, DNSResponse: "192.0.2.42,192.0.2.43", DNSRecord: "CNAME"}
		if !p.Success {
			p.DNSResponse, p.DNSRecord, p.Error = "", "", "Timeout"
		}
		updateImportedSummary(&s, p)
		s.Revision++
		if err := store.Save(s, &p); err != nil {
			t.Fatal(err)
		}
	}
	end := s.Last.Time
	s.EndedAt = &end
	if err := store.Save(s, nil); err != nil {
		t.Fatal(err)
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
	if err != nil || imported.Sent != 4 || imported.Succeeded != 3 || !imported.Last.Time.Equal(s.Last.Time) || imported.Last.DNSResponse != s.Last.DNSResponse {
		t.Fatalf("DNS roundtrip mismatch: %+v %v", imported, err)
	}
	records, err := csv.NewReader(bytes.NewReader(exported.Bytes())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var legacy bytes.Buffer
	w := csv.NewWriter(&legacy)
	for _, row := range records {
		if err := w.Write(row[:len(legacyDNSHeader)]); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	legacyImport, err := store.ImportCSV(bytes.NewReader(legacy.Bytes()))
	if err != nil || legacyImport.Config.TimeoutMS != 4000 || legacyImport.Sent != 4 {
		t.Fatalf("legacy DNS import: %+v %v", legacyImport, err)
	}
	// A bad later row must roll back the whole imported test.
	records[3][3] = "8.8.8.8"
	var invalid bytes.Buffer
	w = csv.NewWriter(&invalid)
	w.WriteAll(records)
	before, _ := store.List("", "stopped", 0)
	if _, err := store.ImportCSV(&invalid); err == nil {
		t.Fatal("accepted changing DNS metadata")
	}
	after, _ := store.List("", "stopped", 0)
	if len(before.Sessions) != len(after.Sessions) {
		t.Fatal("failed import left partial history")
	}
}

func TestDNSProbeParity(t *testing.T) {
	for _, protocol := range []string{"udp", "tcp"} {
		t.Run(protocol, func(t *testing.T) {
			address, requests := dnsFixture(t, protocol)
			for _, test := range []struct{ query, record, response, failure string }{
				{"host.example.test", "A", "192.0.2.42", ""},
				{"alias.example.test", "CNAME", "192.0.2.42", ""},
				{"v6.example.test", "A", "", ""},
				{"missing.example.test", "", "", "Non_Existent_Domain"},
				{"silent.example.test", "", "", "Timeout"},
			} {
				ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
				p := probeDNSWithResolver(ctx, test.query, newDNSResolver(protocol, address))
				cancel()
				if p.Success != (test.failure == "") || p.DNSRecord != test.record || p.DNSResponse != test.response || p.Error != test.failure {
					t.Errorf("%s: %+v", test.query, p)
				}
			}
			if requests.Load() == 0 {
				t.Fatal("custom resolver not used")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			start := time.Now()
			p := probeDNSWithResolver(ctx, "silent.example.test", newDNSResolver(protocol, address))
			if p.Success || time.Since(start) > time.Second {
				t.Fatal("DNS cancellation did not stop promptly")
			}
		})
	}
}

func TestDNSUnrecordedRollbackAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unrecorded.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	r := New(store, nil)
	c, err := validate(dnsConfig())
	if err != nil {
		t.Fatal(err)
	}
	s := Session{ID: "unrecorded", Config: c, Running: true, StartedAt: time.Now(), Revision: 1}
	if err := r.prepareLiveResults(s); err != nil {
		t.Fatal(err)
	}
	p := Sample{Sequence: 1, Time: s.StartedAt, RTT: 4, Success: true, DNSResponse: "192.0.2.42", DNSRecord: "A"}
	updateImportedSummary(&s, p)
	s.Revision++
	if err := r.saveProbe(s, &p); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec("CREATE TRIGGER fail_dns_save BEFORE INSERT ON tests BEGIN SELECT RAISE(ABORT, 'fixture disk failure'); END"); err != nil {
		t.Fatal(err)
	}
	s.Config.Recording = true
	p.Sequence = 2
	p.Time = p.Time.Add(time.Second)
	updateImportedSummary(&s, p)
	s.Revision++
	if err := r.saveProbe(s, &p); err == nil {
		t.Fatal("ignored durable save failure")
	}
	live, err := r.liveResults.Recent(s.ID)
	if err != nil || len(live) != 1 {
		t.Fatalf("failed probe leaked into live chart: %+v %v", live, err)
	}
	if _, err := store.db.Exec("DROP TRIGGER fail_dns_save"); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if saved, err := store.Get(s.ID); err == nil {
		t.Fatalf("unrecorded metadata reached disk: %+v", saved)
	}
	raw, err := store.Recent(s.ID)
	if err != nil || len(raw) != 0 {
		t.Fatalf("unrecorded raw data persisted: %+v %v", raw, err)
	}
}

// Small local wire fixture: no external DNS service or privileged port needed.
func dnsFixture(t *testing.T, protocol string) (string, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	reply := func(query []byte) []byte { requests.Add(1); return dnsReply(query) }
	if protocol == "udp" {
		conn, err := net.ListenPacket("udp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		go func() {
			buffer := make([]byte, 4096)
			for {
				n, peer, err := conn.ReadFrom(buffer)
				if err != nil {
					return
				}
				if response := reply(buffer[:n]); response != nil {
					_, _ = conn.WriteTo(response, peer)
				}
			}
		}()
		return conn.LocalAddr().String(), requests
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(time.Second))
				var size [2]byte
				if _, err := io.ReadFull(conn, size[:]); err != nil {
					return
				}
				query := make([]byte, binary.BigEndian.Uint16(size[:]))
				if _, err := io.ReadFull(conn, query); err != nil {
					return
				}
				response := reply(query)
				if response == nil {
					_, _ = conn.Read(size[:])
					return
				}
				binary.BigEndian.PutUint16(size[:], uint16(len(response)))
				_, _ = conn.Write(append(size[:], response...))
			}()
		}
	}()
	return listener.Addr().String(), requests
}

func dnsReply(query []byte) []byte {
	if len(query) < 17 {
		return nil
	}
	pos := 12
	var labels []string
	for pos < len(query) && query[pos] != 0 {
		n := int(query[pos])
		pos++
		if pos+n >= len(query) {
			return nil
		}
		labels = append(labels, string(query[pos:pos+n]))
		pos += n
	}
	if pos+5 > len(query) {
		return nil
	}
	name := strings.Join(labels, ".")
	if strings.HasPrefix(name, "silent.") {
		return nil
	}
	typeID := binary.BigEndian.Uint16(query[pos+1 : pos+3])
	response := append([]byte{}, query[:pos+5]...)
	binary.BigEndian.PutUint16(response[2:4], 0x8180)
	for i := 6; i < 12; i++ {
		response[i] = 0
	}
	if strings.HasPrefix(name, "missing.") {
		response[3] = 0x83
		return response
	}
	count := uint16(0)
	add := func(kind uint16, data []byte) {
		if strings.HasPrefix(name, "alias.") && kind != 5 {
			response = append(response, []byte{6, 't', 'a', 'r', 'g', 'e', 't', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 4, 't', 'e', 's', 't', 0}...)
		} else {
			response = append(response, 0xc0, 0x0c)
		}
		response = append(response, byte(kind>>8), byte(kind), 0, 1, 0, 0, 0, 60, byte(len(data)>>8), byte(len(data)))
		response = append(response, data...)
		count++
	}
	if strings.HasPrefix(name, "alias.") {
		add(5, []byte{6, 't', 'a', 'r', 'g', 'e', 't', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 4, 't', 'e', 's', 't', 0})
	}
	if typeID == 1 && !strings.HasPrefix(name, "v6.") {
		add(1, []byte{192, 0, 2, 42})
	}
	if typeID == 28 {
		add(28, net.ParseIP("2001:db8::42").To16())
	}
	binary.BigEndian.PutUint16(response[6:8], count)
	return response
}
