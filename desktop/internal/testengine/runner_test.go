package testengine

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func config(url string) Config {
	return Config{URL: url, Method: "GET", IntervalMS: 1000, TimeoutMS: 1000, AcceptedStatuses: []string{"2xx", "3xx"}}
}

func TestCustomExpectedStatuses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	updates := make(chan Session, 10)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	c := config(server.URL)
	c.AcceptedStatuses = []string{"404"}
	s, err := r.Start(c)
	if err != nil {
		t.Fatal(err)
	}
	got := await(t, updates, func(update Session) bool { return update.ID == s.ID && update.Sent == 1 })
	if got.ID != s.ID || !got.Last.Success || got.Succeeded != 1 {
		t.Fatalf("custom expected status was not accepted: %+v", got)
	}
}

func TestProxyAndSecretRedaction(t *testing.T) {
	requests := make(chan *http.Request, 2)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests <- req.Clone(req.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	defer proxy.Close()
	updates := make(chan Session, 20)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	c := config("http://target.invalid/health")
	c.Proxy = ProxyConfig{Enabled: true, URL: proxy.URL, Username: "monitor", Password: "secret"}
	s, err := r.Start(c)
	if err != nil {
		t.Fatal(err)
	}
	if s.Config.Proxy.Password != "" {
		t.Fatal("proxy password exposed in start response")
	}
	got := await(t, updates, func(update Session) bool { return update.ID == s.ID && update.Sent == 1 })
	if !got.Last.Success || got.Last.StatusCode != http.StatusNoContent {
		t.Fatalf("proxy response was not accepted: %+v", got.Last)
	}
	req := <-requests
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("monitor:secret"))
	if req.Header.Get("Proxy-Authorization") != wantAuth {
		t.Fatalf("proxy authorization = %q", req.Header.Get("Proxy-Authorization"))
	}
	if _, err := r.Stop(s.ID); err != nil {
		t.Fatal(err)
	}
	restarted, err := r.Restart(s.ID, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if restarted.ID == s.ID || restarted.Config.Proxy.Password != "" {
		t.Fatalf("unexpected restart response: %+v", restarted)
	}
	await(t, updates, func(update Session) bool { return update.ID == restarted.ID && update.Sent == 1 })
	req = <-requests
	if req.Header.Get("Proxy-Authorization") != wantAuth {
		t.Fatal("restart did not retain the private proxy credential")
	}
}

func await(t *testing.T, updates <-chan Session, accept func(Session) bool) Session {
	t.Helper()
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case s := <-updates:
			if accept(s) {
				return s
			}
		case <-timer.C:
			t.Fatal("timed out waiting for session update")
			return Session{}
		}
	}
}

func TestHTTPResultsAndMethods(t *testing.T) {
	for _, test := range []struct {
		code   int
		method string
	}{{200, "GET"}, {201, "POST"}, {302, "PUT"}, {404, "PATCH"}, {503, "GET"}} {
		t.Run(fmt.Sprintf("%s_%d", test.method, test.code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != test.method {
					t.Errorf("method = %s", req.Method)
				}
				w.Header().Set("Location", "/should-not-follow")
				w.WriteHeader(test.code)
			}))
			defer server.Close()
			updates := make(chan Session, 10)
			r := newTestRunner(t, func(s Session) { updates <- s })
			defer r.Close()
			c := config(server.URL)
			c.Method = test.method
			s, err := r.Start(c)
			if err != nil {
				t.Fatal(err)
			}
			got := await(t, updates, func(s Session) bool { return s.Sent == 1 })
			if got.Last.StatusCode != test.code || got.Last.Success != (test.code < 400) {
				t.Fatalf("unexpected result: %+v", got.Last)
			}
			if got.Succeeded != map[bool]int{true: 1, false: 0}[test.code < 400] {
				t.Fatalf("unexpected statistics: %+v", got)
			}
			r.Stop(s.ID)
			detail, err := r.Get(s.ID)
			if err != nil || len(detail.Samples) != 1 || detail.Session.Running {
				t.Fatalf("unexpected detail: %+v, %v", detail, err)
			}
		})
	}
}

func TestRedirectPolicyAndRestart(t *testing.T) {
	for _, follow := range []bool{false, true} {
		t.Run(fmt.Sprintf("follow_%t", follow), func(t *testing.T) {
			var finalRequests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/final" {
					finalRequests.Add(1)
					if req.Method != http.MethodPost {
						t.Errorf("redirect changed method: %s", req.Method)
					}
					w.WriteHeader(http.StatusCreated)
					return
				}
				http.Redirect(w, req, "/final", http.StatusTemporaryRedirect)
			}))
			defer server.Close()
			updates := make(chan Session, 20)
			r := newTestRunner(t, func(s Session) { updates <- s })
			defer r.Close()
			c := config(server.URL)
			c.Method = http.MethodPost
			c.FollowRedirects = follow
			s, err := r.Start(c)
			if err != nil {
				t.Fatal(err)
			}
			wantCode := http.StatusTemporaryRedirect
			if follow {
				wantCode = http.StatusCreated
			}
			for attempt := 0; attempt < 2; attempt++ {
				got := await(t, updates, func(update Session) bool { return update.ID == s.ID && update.Sent == 1 })
				if !got.Last.Success || got.Last.StatusCode != wantCode || got.Last.ResponsePhase != http.StatusText(wantCode) {
					t.Fatalf("redirect result: %+v", got.Last)
				}
				if _, err := r.Stop(s.ID); err != nil {
					t.Fatal(err)
				}
				if attempt == 0 {
					s, err = r.Restart(s.ID, "")
					if err != nil || s.Config.FollowRedirects != follow || s.Config.Method != http.MethodPost {
						t.Fatalf("restart config: %+v, %v", s, err)
					}
				}
			}
			if got := finalRequests.Load(); (follow && got != 2) || (!follow && got != 0) {
				t.Fatalf("final requests = %d", got)
			}
		})
	}
}

func TestLegacyLongTimings(t *testing.T) {
	c := config("https://example.com")
	c.IntervalMS, c.TimeoutMS = 120000, 90000
	if _, err := validate(c); err != nil {
		t.Fatalf("legacy timings rejected: %v", err)
	}
	c.TimeoutMS = int(^uint(0) >> 1)
	if int64(c.TimeoutMS) > int64((1<<63-1)/time.Millisecond) {
		if _, err := validate(c); err == nil {
			t.Fatal("overflowing timeout accepted")
		}
	}
}

func TestStopCancelsInflightRequest(t *testing.T) {
	entered, cancelled := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		close(entered)
		<-req.Context().Done()
		close(cancelled)
	}))
	defer server.Close()
	r := newTestRunner(t, nil)
	defer r.Close()
	c := config(server.URL)
	c.TimeoutMS = 30000
	s, err := r.Start(c)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request never started")
	}
	stopped, err := r.Stop(s.ID)
	if err != nil || stopped.Running {
		t.Fatalf("stop failed: %v", err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("stop did not cancel HTTP request")
	}
	r.Close()
	detail, _ := r.Get(s.ID)
	if detail.Session.Sent != 0 {
		t.Fatal("user cancellation must not become a failed probe")
	}
	if _, err := r.Start(c); err == nil {
		t.Fatal("closed runner accepted a test")
	}
}

func TestCertificateVerificationAndTimeout(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer tlsServer.Close()
	updates := make(chan Session, 10)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	s, err := r.Start(config(tlsServer.URL))
	if err != nil {
		t.Fatal(err)
	}
	got := await(t, updates, func(s Session) bool { return s.Sent == 1 })
	if got.Last.Success || got.Last.Error == "" {
		t.Fatal("untrusted TLS certificate was accepted")
	}
	r.Stop(s.ID)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { <-req.Context().Done() }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	sample := probe(ctx, server.Client(), config(server.URL))
	if sample.Success || sample.Error == "" {
		t.Fatal("timeout not reported")
	}
}

func TestValidationAndRemoval(t *testing.T) {
	invalidStatus := config("https://example.com")
	invalidStatus.AcceptedStatuses = []string{"199"}
	invalidProxy := config("https://example.com")
	invalidProxy.Proxy = ProxyConfig{Enabled: true, URL: "socks5://proxy.example:1080"}
	for _, c := range []Config{config("file:///tmp/test"), config("http://"), config("https://user:pass@example.com"), config("http://example.com:70000"), {URL: "https://example.com", Method: "DELETE", IntervalMS: 1000, TimeoutMS: 1000, AcceptedStatuses: []string{"2xx"}}, {URL: "https://example.com", Method: "GET", IntervalMS: 0, TimeoutMS: 1000, AcceptedStatuses: []string{"2xx"}}, invalidStatus, invalidProxy} {
		if _, err := validate(c); err == nil {
			t.Errorf("accepted invalid config: %+v", c)
		}
	}
	r := newTestRunner(t, nil)
	defer r.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(200) }))
	defer server.Close()
	s, err := r.Start(config(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(s.ID); err == nil {
		t.Fatal("removed active test")
	}
	r.Stop(s.ID)
	if err := r.Remove(s.ID); err != nil {
		t.Fatal(err)
	}
	page, err := r.List("", "all", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Sessions) != 0 {
		t.Fatal("removed test remains in list")
	}
}

func newTestRunner(t *testing.T, callback func(Session)) *Runner {
	t.Helper()
	store, err := OpenStore(filepath.Join(t.TempDir(), "results.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return New(store, callback)
}

func TestCurrentSessionRetainsStoppedResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	updates := make(chan Session, 20)
	r := newTestRunner(t, func(s Session) { updates <- s })
	defer r.Close()
	first, err := r.Start(config(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	await(t, updates, func(s Session) bool { return s.ID == first.ID && s.Sent > 0 })
	stopped, err := r.Stop(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.Restart(first.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	// Switching tabs repeatedly must retain the first run even after worker cleanup.
	for _, filter := range []string{"current", "stopped", "current"} {
		page, err := r.List("", filter, 0)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, s := range page.Sessions {
			if s.ID == first.ID {
				found = true
				if s.Running || s.Sent != stopped.Sent || s.Last == nil || s.Last.StatusCode != http.StatusNoContent {
					t.Fatalf("stopped results changed: %+v", s)
				}
			}
		}
		if !found || page.Active != 1 {
			t.Fatalf("%s page lost stopped test or active count: %+v", filter, page)
		}
	}
	if _, err := r.Stop(second.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove(first.ID); err != nil {
		t.Fatal(err)
	}
	for _, filter := range []string{"current", "stopped"} {
		page, err := r.List("", filter, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Sessions) != 1 || page.Sessions[0].ID != second.ID {
			t.Fatalf("deletion not reflected in %s: %+v", filter, page)
		}
	}
}

func TestTenSharedSlotsAndBatchCapacity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	defer server.Close()
	r := newTestRunner(t, nil)
	defer r.Close()
	c, err := validate(config(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	var first Session
	for i := 0; i < 9; i++ {
		s, err := r.Start(c)
		if err != nil {
			t.Fatalf("slot %d: %v", i+1, err)
		}
		if i == 0 {
			first = s
		}
	}
	if _, err := r.startBatch([]Config{c, c}); err == nil {
		t.Fatal("oversized batch accepted")
	}
	page, err := r.List("", "current", 0)
	if err != nil || page.Active != 9 {
		t.Fatalf("batch partially started: %+v, %v", page, err)
	}
	if _, err := r.startBatch([]Config{c}); err != nil {
		t.Fatalf("tenth slot rejected: %v", err)
	}
	if _, err := r.Start(c); err == nil {
		t.Fatal("eleventh test accepted")
	}
	page, err = r.List("no-match", "stopped-dns", 1)
	if err != nil || page.Capacity != 10 || page.Active != 10 || page.ActiveByType["http"] != 10 {
		t.Fatalf("incorrect global pool: %+v, %v", page, err)
	}
	if _, err := r.Stop(first.ID); err != nil {
		t.Fatal(err)
	}
	page, err = r.List("", "current", 0)
	if err != nil || page.Active != 9 || page.ActiveByType["http"] != 9 {
		t.Fatalf("stop did not free slot: %+v, %v", page, err)
	}
	if _, err := r.Restart(first.ID, ""); err != nil {
		t.Fatalf("released slot could not be reused: %v", err)
	}
}

func TestPoolCountsAllTypesIndependentOfVisibleRows(t *testing.T) {
	r := newTestRunner(t, nil)
	// Lifecycle snapshots isolate aggregation from network availability and ICMP permissions.
	for i, kind := range []string{"", "http", "dns", "tcp", "tcp", "icmp"} {
		id := fmt.Sprint(i)
		r.sessions[id] = &state{Session: Session{ID: id, Config: Config{Type: kind}, Running: true}}
	}
	r.sessions["stopped"] = &state{Session: Session{Config: Config{Type: "dns"}, Running: false}}
	for _, filter := range []string{"current-http", "current-dns", "current-tcp", "current-icmp", "stopped", "stopped-tcp"} {
		page, err := r.List("no-match", filter, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Sessions) != 0 || page.Active != 6 || page.Capacity != 10 {
			t.Fatalf("%s: %+v", filter, page)
		}
		for kind, want := range map[string]int{"http": 2, "dns": 1, "tcp": 2, "icmp": 1} {
			if page.ActiveByType[kind] != want {
				t.Fatalf("%s: %s count = %d, want %d", filter, kind, page.ActiveByType[kind], want)
			}
		}
	}
	// A terminal storage failure releases capacity just like an explicit stop.
	r.sessions["5"].Running = false
	r.sessions["5"].EndReason = "storage_error"
	page, err := r.List("", "stopped", 0)
	if err != nil || page.Active != 5 || page.ActiveByType["icmp"] != 0 {
		t.Fatalf("terminal state retained slot: %+v, %v", page, err)
	}
}
