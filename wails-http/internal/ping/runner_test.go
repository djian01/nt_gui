package ping

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func config(url string) Config {
	return Config{URL: url, Method: "GET", IntervalMS: 1000, TimeoutMS: 1000}
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

func TestHTTPResults(t *testing.T) {
	for _, code := range []int{200, 302, 404, 503} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != "HEAD" {
					t.Errorf("method = %s", req.Method)
				}
				w.Header().Set("Location", "/should-not-follow")
				w.WriteHeader(code)
			}))
			defer server.Close()
			updates := make(chan Session, 10)
			r := New(func(s Session) { updates <- s })
			defer r.Close()
			c := config(server.URL)
			c.Method = "HEAD"
			s, err := r.Start(c)
			if err != nil {
				t.Fatal(err)
			}
			got := await(t, updates, func(s Session) bool { return s.Sent == 1 })
			if got.Last.StatusCode != code || got.Last.Success != (code < 400) {
				t.Fatalf("unexpected result: %+v", got.Last)
			}
			if got.Succeeded != map[bool]int{true: 1, false: 0}[code < 400] {
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

func TestStopCancelsInflightRequest(t *testing.T) {
	entered, cancelled := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		close(entered)
		<-req.Context().Done()
		close(cancelled)
	}))
	defer server.Close()
	r := New(nil)
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
	r := New(func(s Session) { updates <- s })
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
	for _, c := range []Config{config("file:///tmp/test"), config("http://"), config("https://user:pass@example.com"), config("http://example.com:70000"), {URL: "https://example.com", Method: "POST", IntervalMS: 1000, TimeoutMS: 1000}, {URL: "https://example.com", Method: "GET", IntervalMS: 0, TimeoutMS: 1000}} {
		if _, err := validate(c); err == nil {
			t.Errorf("accepted invalid config: %+v", c)
		}
	}
	r := New(nil)
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
	if len(r.List()) != 0 {
		t.Fatal("removed test remains in list")
	}
}

func TestBoundedHistoryOrderingAndIsolation(t *testing.T) {
	r := New(nil)
	s := &state{Session: Session{ID: "test"}, samples: make([]Sample, MaxSamples), next: 3}
	for i := 0; i < MaxSamples; i++ {
		s.samples[(i+3)%MaxSamples] = Sample{Sequence: i + 4}
	}
	r.sessions[s.ID] = s
	detail, err := r.Get(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Samples) != MaxSamples || detail.Samples[0].Sequence != 4 || detail.Samples[MaxSamples-1].Sequence != MaxSamples+3 {
		t.Fatal("ring buffer returned out-of-order data")
	}
	detail.Samples[0].Sequence = -1
	if s.samples[3].Sequence == -1 {
		t.Fatal("snapshot aliases mutable history")
	}
}
