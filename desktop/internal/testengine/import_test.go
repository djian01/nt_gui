package testengine

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestImportWailsCSVRoundTrip(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	start := time.Date(2026, 9, 8, 1, 2, 3, 456000000, time.UTC)
	original := Session{ID: "original", Config: config("https://example.com/health?q=1"), StartedAt: start, Revision: 1}
	original.Config.Method = "POST"
	original.Config.IntervalMS = 120000
	original.Config.TimeoutMS = 90000
	original.Config.FollowRedirects = true
	original.Config.Proxy = ProxyConfig{Enabled: true, URL: "http://proxy.example:8080", Username: "tester"}
	original.PasswordRequired = true
	for i, sample := range []Sample{
		{Sequence: 1, Time: start, RTT: 12.25, StatusCode: 204, Success: true},
		{Sequence: 2, Time: start.Add(time.Second), RTT: 31.5, StatusCode: 503, Error: "HTTP 503 Service Unavailable"},
		{Sequence: 3, Time: start.Add(2 * time.Second), RTT: 14.75, StatusCode: 204, Success: true},
	} {
		updateImportedSummary(&original, sample)
		original.Revision = i + 2
		if err := store.Save(original, &sample); err != nil {
			t.Fatal(err)
		}
	}
	var exported bytes.Buffer
	if err := store.ExportCSV(original.ID, &exported); err != nil {
		t.Fatal(err)
	}
	imported, err := store.ImportCSV(bytes.NewReader(exported.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if imported.ID == original.ID || imported.Running || imported.EndReason != "imported" {
		t.Fatalf("imported identity/state is wrong: %+v", imported)
	}
	if imported.Sent != 3 || imported.Succeeded != 2 || imported.MinRTT != 12.25 || imported.MaxRTT != 14.75 || imported.AvgRTT != 13.5 {
		t.Fatalf("imported summary differs: %+v", imported)
	}
	if imported.Config.URL != original.Config.URL || imported.Config.Method != original.Config.Method || imported.Config.IntervalMS != original.Config.IntervalMS || imported.Config.TimeoutMS != original.Config.TimeoutMS {
		t.Fatalf("imported metadata differs: %+v", imported.Config)
	}
	if !reflect.DeepEqual(imported.Config, original.Config) || !imported.PasswordRequired {
		t.Fatalf("replay metadata differs: %+v", imported.Config)
	}
	detail, err := store.Timeline(imported.ID, 0, 0)
	if err != nil || detail.Count != 3 || detail.Succeeded != 2 || len(detail.Samples) != 3 {
		t.Fatalf("imported timeline differs: %+v, %v", detail, err)
	}
}

func TestImportLegacyHTTPWithSummaryMetadata(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var input bytes.Buffer
	w := csv.NewWriter(&input)
	if err := w.Write(legacyHTTPHeader); err != nil {
		t.Fatal(err)
	}
	rows := [][]string{
		{"http", "0", "true", "POST", "https://example.com:8443/path", "201", "Created", "10.5", "2026-09-08", "11:00:00 AEST", "1", "1", "0.00%", "10.5ms", "10.5ms", "10.5ms", ""},
		{"http", "1", "false", "POST", "https://example.com:8443/path", "500", "Server Error", "20", "2026-09-08", "11:00:01 AEST", "2", "1", "50.00%", "10.5ms", "10.5ms", "10.5ms", "upstream failed"},
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	imported, err := store.ImportCSV(&input)
	if err != nil {
		t.Fatal(err)
	}
	if imported.Config.Method != "POST" || imported.Sent != 2 || imported.Succeeded != 1 || imported.MinRTT != 10.5 || imported.AvgRTT != 10.5 || imported.MaxRTT != 10.5 {
		t.Fatalf("legacy metadata differs: %+v", imported)
	}
	detail, err := store.Recent(imported.ID)
	if err != nil || len(detail) != 2 {
		t.Fatalf("legacy probes missing: %+v, %v", detail, err)
	}
	if detail[0].Sequence != 1 || detail[1].Sequence != 2 || detail[1].Error != "upstream failed" || detail[1].StatusCode != 500 {
		t.Fatalf("legacy samples differ: %+v", detail)
	}
	if detail[0].ResponsePhase != "Created" || detail[0].Time.UTC().Hour() != 1 || imported.ImportNote == "" || !imported.Config.FollowRedirects {
		t.Fatalf("legacy context lost: %+v, %+v", imported, detail[0])
	}
}

func TestImportCSVIsAtomicOnInvalidRow(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	input := strings.Join([]string{
		strings.Join(wailsCSVHeader, ","),
		"source,https://example.com,GET,1000,4000,1,2026-09-08T00:00:00Z,12,200,true,",
		"source,https://example.com,GET,1000,4000,3,2026-09-08T00:00:01Z,13,200,true,",
	}, "\n")
	if _, err := store.ImportCSV(strings.NewReader(input)); err == nil {
		t.Fatal("invalid sequence was accepted")
	}
	var tests, samples, buckets int
	for table, target := range map[string]*int{"tests": &tests, "samples": &samples, "buckets": &buckets} {
		if err := store.db.QueryRow("SELECT count(*) FROM " + table).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if tests != 0 || samples != 0 || buckets != 0 {
		t.Fatalf("failed import committed data: tests=%d samples=%d buckets=%d", tests, samples, buckets)
	}
}

func TestImportLegacyPartialRecording(t *testing.T) {
	store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var input bytes.Buffer
	w := csv.NewWriter(&input)
	_ = w.Write(legacyHTTPHeader)
	_ = w.Write([]string{"http", "10", "true", "GET", "https://example.com", "200", "OK", "50", "2026-09-08", "11:00:00 AEST", "11", "10", "9.09%", "1ms", "10ms", "50ms", ""})
	_ = w.Write([]string{"http", "11", "false", "GET", "https://example.com", "503", "Unavailable", "80", "2026-09-08", "11:00:01 AEST", "12", "10", "16.67%", "1ms", "10ms", "50ms", "unavailable"})
	w.Flush()
	got, err := store.ImportCSV(&input)
	if err != nil {
		t.Fatal(err)
	}
	if got.Sent != 2 || got.Succeeded != 1 || got.MinRTT != 50 || got.AvgRTT != 50 || !strings.Contains(got.ImportNote, "#10") {
		t.Fatalf("incorrect imported subset: %+v", got)
	}
	timeline, err := store.Timeline(got.ID, 0, 0)
	if err != nil || timeline.Count != 2 || timeline.Average != got.AvgRTT || timeline.Samples[0].Sequence != 1 {
		t.Fatalf("subset timeline: %+v, %v", timeline, err)
	}
}

func TestImportRejectsChangedMetadataAndCredentials(t *testing.T) {
	for _, badProxy := range []string{"http://user:secret@proxy.example:8080", "http://different.example:8080"} {
		t.Run(badProxy, func(t *testing.T) {
			store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			var input bytes.Buffer
			w := csv.NewWriter(&input)
			_ = w.Write(wailsCSVFullHeader)
			row := []string{"source", "https://example.com", "POST", "1000", "4000", "1", "2026-09-08T00:00:00Z", "5", "200", "true", "", "OK", "2xx,3xx", "http://proxy.example:8080", "tester", "true", "false"}
			_ = w.Write(row)
			row[5], row[6], row[13] = "2", "2026-09-08T00:00:01Z", badProxy
			_ = w.Write(row)
			w.Flush()
			if _, err := store.ImportCSV(&input); err == nil {
				t.Fatal("invalid proxy metadata accepted")
			}
			page, err := store.List("", "all", 0)
			if err != nil || len(page.Sessions) != 0 {
				t.Fatalf("partial import leaked: %+v, %v", page, err)
			}
		})
	}
}

func TestRunnerImportDoesNotBlockLiveProbes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	defer server.Close()
	r := newTestRunner(t, nil)
	defer r.Close()
	if _, err := r.Start(config(server.URL)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ImportCSV(strings.NewReader("")); err == nil || !strings.Contains(err.Error(), "Stop active tests") {
		t.Fatalf("active import not rejected: %v", err)
	}
}
