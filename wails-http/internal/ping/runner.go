// Package ping owns HTTP sessions independently of any desktop toolkit.
package ping

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MaxActive = 8
)

type Config struct {
	URL              string      `json:"url"`
	Method           string      `json:"method"`
	IntervalMS       int         `json:"intervalMs"`
	TimeoutMS        int         `json:"timeoutMs"`
	AcceptedStatuses []string    `json:"acceptedStatuses"`
	Proxy            ProxyConfig `json:"proxy"`

	accepted []statusRange
	proxyURL *url.URL
}

type ProxyConfig struct {
	Enabled  bool   `json:"enabled"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type statusRange struct {
	min int
	max int
}

type Sample struct {
	Sequence   int       `json:"sequence"`
	Time       time.Time `json:"time"`
	RTT        float64   `json:"rtt"`
	StatusCode int       `json:"statusCode"`
	Success    bool      `json:"success"`
	Error      string    `json:"error"`
}

type Session struct {
	ID               string     `json:"id"`
	Config           Config     `json:"config"`
	Running          bool       `json:"running"`
	StartedAt        time.Time  `json:"startedAt"`
	EndedAt          *time.Time `json:"endedAt"`
	Revision         int        `json:"revision"`
	EndReason        string     `json:"endReason"`
	SaveError        string     `json:"saveError"`
	PasswordRequired bool       `json:"passwordRequired"`
	Sent             int        `json:"sent"`
	Succeeded        int        `json:"succeeded"`
	MinRTT           float64    `json:"minRtt"`
	MaxRTT           float64    `json:"maxRtt"`
	AvgRTT           float64    `json:"avgRtt"`
	Last             *Sample    `json:"last"`
}

type Detail struct {
	Session Session  `json:"session"`
	Samples []Sample `json:"samples"`
}

type state struct {
	Session
	requestConfig Config
	cancel        context.CancelFunc
}

type Runner struct {
	mu       sync.Mutex
	sessions map[string]*state
	store    *Store
	closed   bool
	wg       sync.WaitGroup
	onUpdate func(Session)
}

func New(store *Store, onUpdate func(Session)) *Runner {
	return &Runner{sessions: make(map[string]*state), store: store, onUpdate: onUpdate}
}

func validate(c Config) (Config, error) {
	c.URL = strings.TrimSpace(c.URL)
	u, err := url.Parse(c.URL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return c, errors.New("Enter a complete http:// or https:// URL")
	}
	if len(c.URL) > 4096 || u.User != nil || u.Fragment != "" {
		return c, errors.New("URL must be under 4096 characters, without credentials or a fragment")
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return c, errors.New("Port must be between 1 and 65535")
		}
	}
	c.Method = strings.ToUpper(strings.TrimSpace(c.Method))
	if c.Method != http.MethodGet && c.Method != http.MethodPut && c.Method != http.MethodPatch {
		return c, errors.New("Choose GET, PUT, or PATCH")
	}
	if c.IntervalMS < 1000 || c.IntervalMS > 60000 {
		return c, errors.New("Interval must be between 1 and 60 seconds")
	}
	if c.TimeoutMS < 1000 || c.TimeoutMS > 30000 {
		return c, errors.New("Timeout must be between 1 and 30 seconds")
	}
	c.AcceptedStatuses, c.accepted, err = validateStatuses(c.AcceptedStatuses)
	if err != nil {
		return c, err
	}
	c.Proxy.URL = strings.TrimSpace(c.Proxy.URL)
	c.Proxy.Username = strings.TrimSpace(c.Proxy.Username)
	if c.Proxy.Enabled {
		if len(c.Proxy.URL) > 2048 || len(c.Proxy.Username) > 256 || len(c.Proxy.Password) > 1024 {
			return c, errors.New("Proxy details are too long")
		}
		proxyURL, parseErr := url.Parse(c.Proxy.URL)
		if parseErr != nil || proxyURL.Hostname() == "" || (proxyURL.Scheme != "http" && proxyURL.Scheme != "https") {
			return c, errors.New("Enter a complete http:// or https:// proxy URL")
		}
		if proxyURL.User != nil || proxyURL.Path != "" || proxyURL.RawQuery != "" || proxyURL.Fragment != "" {
			return c, errors.New("Proxy URL cannot include credentials, a path, query, or fragment")
		}
		if port := proxyURL.Port(); port != "" {
			n, portErr := strconv.Atoi(port)
			if portErr != nil || n < 1 || n > 65535 {
				return c, errors.New("Proxy port must be between 1 and 65535")
			}
		}
		if c.Proxy.Username == "" && c.Proxy.Password != "" {
			return c, errors.New("Enter a proxy username when a password is provided")
		}
		if c.Proxy.Username != "" {
			proxyURL.User = url.UserPassword(c.Proxy.Username, c.Proxy.Password)
		}
		c.proxyURL = proxyURL
	} else {
		c.Proxy = ProxyConfig{}
	}
	return c, nil
}

func validateStatuses(values []string) ([]string, []statusRange, error) {
	if len(values) == 0 || len(values) > 20 {
		return nil, nil, errors.New("Choose at least one expected HTTP status, up to 20")
	}
	normalized := make([]string, 0, len(values))
	ranges := make([]statusRange, 0, len(values))
	seen := make(map[string]bool)
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if seen[value] {
			continue
		}
		var current statusRange
		if len(value) == 3 && value[1:] == "xx" && value[0] >= '2' && value[0] <= '5' {
			current.min = int(value[0]-'0') * 100
			current.max = current.min + 99
		} else {
			code, parseErr := strconv.Atoi(value)
			if parseErr != nil || code < 200 || code > 599 {
				return nil, nil, fmt.Errorf("Expected HTTP status %q must be 2xx–5xx or a code from 200 to 599", value)
			}
			current = statusRange{min: code, max: code}
			value = strconv.Itoa(code)
		}
		seen[value] = true
		normalized = append(normalized, value)
		ranges = append(ranges, current)
	}
	return normalized, ranges, nil
}

func (r *Runner) Start(config Config) (Session, error) {
	c, err := validate(config)
	if err != nil {
		return Session{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.startLocked(c)
}

func (r *Runner) startLocked(c Config) (Session, error) {
	if r.closed {
		return Session{}, errors.New("Application is closing")
	}
	active := 0
	for id, s := range r.sessions {
		if !s.Running {
			delete(r.sessions, id)
			continue
		}
		if s.Running {
			active++
		}
	}
	if active >= MaxActive {
		return Session{}, errors.New("Stop a test before starting another (8 active tests maximum)")
	}
	ctx, cancel := context.WithCancel(context.Background())
	publicConfig := c
	publicConfig.Proxy.Password = ""
	publicConfig.accepted = nil
	publicConfig.proxyURL = nil
	s := &state{Session: Session{ID: rand.Text(), Config: publicConfig, Running: true, StartedAt: time.Now(), Revision: 1, PasswordRequired: c.Proxy.Password != ""}, requestConfig: c, cancel: cancel}
	if err := r.store.Save(s.Session, nil); err != nil {
		cancel()
		return Session{}, fmt.Errorf("Could not save the new test: %w", err)
	}
	r.sessions[s.ID] = s
	r.wg.Add(1)
	go r.run(ctx, s)
	return s.Session, nil
}

func (r *Runner) Restart(id, password string) (Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, err := r.store.Get(id)
	if err != nil {
		return Session{}, err
	}
	if old.Running {
		return Session{}, errors.New("Stop the test before running it again")
	}
	if old.PasswordRequired && password == "" {
		return Session{}, errors.New("Enter the proxy password to run this saved test again")
	}
	old.Config.Proxy.Password = password
	c, err := validate(old.Config)
	if err != nil {
		return Session{}, err
	}
	return r.startLocked(c)
}

// emitLocked preserves revision order; callbacks must not call back into Runner.
func (r *Runner) emitLocked(s *state) {
	if r.onUpdate != nil {
		r.onUpdate(s.Session)
	}
}

func (r *Runner) run(ctx context.Context, s *state) {
	defer r.wg.Done()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Fresh connections and normal system certificate verification.
	if s.requestConfig.proxyURL == nil {
		transport.Proxy = nil
	} else {
		transport.Proxy = http.ProxyURL(s.requestConfig.proxyURL)
	}
	transport.DisableKeepAlives = true
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	r.mu.Lock()
	r.emitLocked(s)
	r.mu.Unlock()
	for {
		started := time.Now()
		requestCtx, cancel := context.WithTimeout(ctx, time.Duration(s.Config.TimeoutMS)*time.Millisecond)
		sample := probe(requestCtx, client, s.requestConfig)
		cancel()
		r.mu.Lock()
		if ctx.Err() != nil || !s.Running {
			r.mu.Unlock()
			return
		}
		next := s.Session
		next.Sent++
		sample.Sequence = next.Sent
		next.Last = &sample
		if sample.Success {
			next.Succeeded++
			if next.Succeeded == 1 || sample.RTT < next.MinRTT {
				next.MinRTT = sample.RTT
			}
			next.MaxRTT = max(next.MaxRTT, sample.RTT)
			next.AvgRTT += (sample.RTT - next.AvgRTT) / float64(next.Succeeded)
		}
		next.Revision++
		if err := r.store.Save(next, &sample); err != nil {
			s.cancel()
			s.Running = false
			now := time.Now()
			s.EndedAt = &now
			s.EndReason = "storage_error"
			s.SaveError = "Saving failed; this test has stopped. " + err.Error()
			s.Revision++
			_ = r.store.Save(s.Session, nil)
			r.emitLocked(s)
			r.mu.Unlock()
			return
		}
		s.Session = next
		r.emitLocked(s)
		r.mu.Unlock()
		// Start-to-start interval; never overlap requests for the same session.
		timer := time.NewTimer(max(0, time.Duration(s.Config.IntervalMS)*time.Millisecond-time.Since(started)))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func probe(ctx context.Context, client *http.Client, c Config) Sample {
	s := Sample{Time: time.Now()}
	req, err := http.NewRequestWithContext(ctx, c.Method, c.URL, nil)
	if err != nil {
		s.Error = err.Error()
		return s
	}
	req.Header.Set("User-Agent", "NTGUI-HTTP-Prototype/0.1")
	resp, err := client.Do(req)
	s.RTT = float64(time.Since(s.Time).Microseconds()) / 1000
	if err != nil {
		// Do not reflect a URL with query tokens back into error messages.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		s.Error = err.Error()
		return s
	}
	resp.Body.Close()
	s.StatusCode = resp.StatusCode
	for _, accepted := range c.accepted {
		if resp.StatusCode >= accepted.min && resp.StatusCode <= accepted.max {
			s.Success = true
			break
		}
	}
	if !s.Success {
		s.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return s
}

func (r *Runner) List(search, filter string, before int64) (Page, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	page, err := r.store.List(search, filter, before)
	for _, s := range r.sessions {
		if s.Running {
			page.Active++
		}
	}
	for i, s := range page.Sessions {
		if current := r.sessions[s.ID]; current != nil {
			page.Sessions[i] = current.Session
		}
	}
	return page, err
}

func (r *Runner) Get(id string) (Detail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, err := r.store.Get(id)
	if err != nil {
		return Detail{}, err
	}
	if current := r.sessions[id]; current != nil {
		session = current.Session
	}
	samples, err := r.store.Recent(id)
	return Detail{Session: session, Samples: samples}, err
}

func (r *Runner) Stop(id string) (Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return r.store.Get(id)
	}
	s.cancel()
	if s.Running {
		s.Running = false
		now := time.Now()
		s.EndedAt = &now
		s.EndReason = "stopped"
		s.Revision++
	}
	err := r.store.Save(s.Session, nil)
	if err != nil {
		s.SaveError = "Could not save the stopped state: " + err.Error()
	}
	r.emitLocked(s)
	return s.Session, err
}

func (r *Runner) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s := r.sessions[id]; s != nil && s.Running {
		return errors.New("Stop the test before deleting its history")
	}
	if err := r.store.Remove(id); err != nil {
		return err
	}
	delete(r.sessions, id)
	return nil
}

func (r *Runner) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	for _, s := range r.sessions {
		s.cancel()
	}
	var saveErrors []error
	for _, s := range r.sessions {
		if s.Running {
			s.Running = false
			now := time.Now()
			s.EndedAt = &now
			s.EndReason = "stopped"
			s.Revision++
		}
		if err := r.store.Save(s.Session, nil); err != nil {
			saveErrors = append(saveErrors, err)
		}
	}
	r.mu.Unlock()
	r.wg.Wait()
	return errors.Join(saveErrors...)
}
