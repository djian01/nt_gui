package testengine

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestUnrecordedLifecycleNeverWritesDisk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	defer server.Close()
	for _, protocol := range []string{"http", "dns"} {
		t.Run(protocol, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "results.db")
			store, err := OpenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			// Any attempted durable write now fails. A temporary test must
			// still support its whole lifecycle, including app shutdown.
			if _, err := store.db.Exec("PRAGMA query_only=ON"); err != nil {
				t.Fatal(err)
			}
			updates := make(chan Session, 30)
			r := New(store, func(s Session) { updates <- s })
			defer r.Close()
			c := config(server.URL)
			c.Type = "http"
			if protocol == "dns" {
				c = dnsConfig()
			}
			c.Recording = false
			s, err := r.Start(c)
			if err != nil {
				t.Fatal(err)
			}
			await(t, updates, func(s Session) bool { return s.Sent == 1 })
			if detail, err := r.Get(s.ID); err != nil || len(detail.Samples) != 1 {
				t.Fatalf("temporary detail missing: %+v %v", detail, err)
			}
			if _, err := store.Get(s.ID); err == nil {
				t.Fatal("unrecorded metadata saved")
			}
			if _, err := r.Stop(s.ID); err != nil {
				t.Fatal(err)
			}
			if history, err := r.List("", "stopped", 0); err != nil || len(history.Sessions) != 0 {
				t.Fatalf("temporary test in History: %+v %v", history, err)
			}
			if current, err := r.List("", "current-"+protocol, 0); err != nil || len(current.Sessions) != 1 {
				t.Fatalf("temporary test missing from current list: %+v %v", current, err)
			}
			replayed, err := r.Restart(s.ID, "")
			if err != nil || recordingEnabled(replayed.Config) {
				t.Fatalf("temporary replay: %+v %v", replayed, err)
			}
			if _, err := r.Stop(replayed.ID); err != nil {
				t.Fatal(err)
			}
			if err := r.Remove(s.ID); err != nil {
				t.Fatal(err)
			}
			if protocol == "dns" {
				err = r.Dismiss(replayed.ID)
			} else {
				err = r.Remove(replayed.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := r.Start(c); err != nil {
				t.Fatal(err)
			}
			if err := r.Close(); err != nil {
				t.Fatal(err)
			}
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := OpenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			for _, table := range []string{"tests", "samples", "buckets"} {
				var count int
				if err := reopened.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s contains temporary data: %d %v", table, count, err)
				}
			}
		})
	}
}

func TestHTTPOptionalRecording(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	updates := make(chan Session, 30)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	c := config(server.URL)
	c.Type = "http"
	c.Recording = false
	s, err := r.Start(c)
	if err != nil {
		t.Fatal(err)
	}
	first := await(t, updates, func(s Session) bool { return s.Sent == 1 })
	if !first.Last.Success || first.Config.Recording {
		t.Fatalf("unexpected unrecorded HTTP result: %+v", first)
	}
	if raw, err := r.store.Recent(s.ID); err != nil || len(raw) != 0 {
		t.Fatalf("unrecorded HTTP probes persisted: %+v %v", raw, err)
	}
	if _, err := r.Record(s.ID); err != nil {
		t.Fatal(err)
	}
	await(t, updates, func(s Session) bool { return s.Sent == 2 })
	if _, err := r.Stop(s.ID); err != nil {
		t.Fatal(err)
	}
	if raw, err := r.store.Recent(s.ID); err != nil || len(raw) != 1 || raw[0].Sequence != 2 {
		t.Fatalf("wrong recorded boundary: %+v %v", raw, err)
	}
	if live, err := r.Timeline(s.ID, 0, 0); err != nil || live.Count != 2 {
		t.Fatalf("live HTTP chart lost samples: %+v %v", live, err)
	}
	var output bytes.Buffer
	if err := r.store.ExportCSV(s.ID, &output); err != nil {
		t.Fatal(err)
	}
	imported, err := r.ImportCSV(&output)
	if err != nil || imported.Sent != 1 || imported.Succeeded != 1 || imported.Config.URL != c.URL {
		t.Fatalf("partial HTTP recording import: %+v %v", imported, err)
	}
	restarted, err := r.Restart(s.ID, "")
	if err != nil || !recordingEnabled(restarted.Config) {
		t.Fatalf("HTTP replay lost recording state: %+v %v", restarted, err)
	}
	if _, err := r.Stop(restarted.ID); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPLegacyRecordingCompatibility(t *testing.T) {
	// Missing type is the old HTTP-only API/database format. Even snapshots
	// with recording:false were recorded before HTTP acquired this option.
	c := config("https://example.com")
	if !recordingEnabled(c) {
		t.Fatal("legacy HTTP recording became unavailable")
	}
	validated, err := validate(c)
	if err != nil || !recordingEnabled(validated) {
		t.Fatalf("legacy HTTP replay stopped recording: %+v %v", validated, err)
	}
	c.Type = "http"
	validated, err = validate(c)
	if err != nil || recordingEnabled(validated) {
		t.Fatalf("explicit recording off ignored: %+v %v", validated, err)
	}
	c.Recording = true
	validated, err = validate(c)
	if err != nil || !recordingEnabled(validated) {
		t.Fatalf("explicit recording on ignored: %+v %v", validated, err)
	}
}
