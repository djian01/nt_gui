package testengine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func validateTCP(c Config) (Config, error) {
	c.Target = strings.TrimSpace(c.Target)
	if net.ParseIP(c.Target) == nil {
		name := strings.TrimSuffix(c.Target, ".")
		if len(name) == 0 || len(name) > 253 {
			return c, errors.New("Enter a TCP server IP address or hostname")
		}
		for _, label := range strings.Split(name, ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return c, errors.New("Invalid TCP hostname")
			}
			for _, ch := range label {
				if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '-') {
					return c, errors.New("Enter a hostname or IP without a scheme, port, or path")
				}
			}
		}
	}
	if c.Port < 1 || c.Port > 65535 {
		return c, errors.New("TCP port must be between 1 and 65535")
	}
	const maxDurationMS = int64((1<<63 - 1) / time.Millisecond)
	if c.IntervalMS < 1000 || c.IntervalMS%1000 != 0 || int64(c.IntervalMS) > maxDurationMS {
		return c, errors.New("Interval must be a positive whole number of seconds")
	}
	if c.TimeoutMS < 1000 || c.TimeoutMS%1000 != 0 || int64(c.TimeoutMS) > maxDurationMS {
		return c, errors.New("Timeout must be a positive whole number of seconds")
	}
	if c.ResolvedIP != "" && net.ParseIP(c.ResolvedIP) == nil {
		return c, errors.New("Invalid resolved TCP IP address")
	}
	return Config{Type: "tcp", Target: c.Target, Port: c.Port, ResolvedIP: c.ResolvedIP,
		URL: net.JoinHostPort(c.Target, strconv.Itoa(c.Port)), Recording: c.Recording,
		IntervalMS: c.IntervalMS, TimeoutMS: c.TimeoutMS, AcceptedStatuses: []string{}}, nil
}

func resolveTCP(c Config) (Config, error) {
	if c.ResolvedIP != "" {
		return c, nil
	}
	if ip := net.ParseIP(c.Target); ip != nil {
		c.ResolvedIP = ip.String()
		return c, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), min(10*time.Second, time.Duration(c.TimeoutMS)*time.Millisecond))
	defer cancel()
	addresses, err := net.DefaultResolver.LookupHost(ctx, c.Target)
	if err != nil {
		return c, fmt.Errorf("Cannot resolve TCP target %s: %w", c.Target, err)
	}
	if len(addresses) == 0 {
		return c, fmt.Errorf("No IP address for TCP target %s", c.Target)
	}
	c.ResolvedIP = addresses[0]
	return c, nil
}

// StartTCP resolves all targets before starting any connection probes.
func (r *Runner) StartTCP(config Config, targets string) ([]Session, error) {
	if len(targets) > 65536 {
		return nil, errors.New("TCP target list is too long")
	}
	var configs []Config
	for _, line := range strings.Split(targets, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		config.Type, config.Target, config.ResolvedIP = "tcp", strings.TrimSpace(line), ""
		c, err := validateTCP(config)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
		if len(configs) > MaxActive {
			return nil, fmt.Errorf("%d active tests maximum", MaxActive)
		}
	}
	if len(configs) == 0 {
		return nil, errors.New("Enter at least one TCP server IP or hostname")
	}
	for i, c := range configs {
		resolved, err := resolveTCP(c)
		if err != nil {
			return nil, err
		}
		configs[i] = resolved
	}
	return r.startBatch(configs)
}

// Only the connection handshake is measured; no application payload is sent.
func probeTCP(ctx context.Context, c Config) Sample {
	s := Sample{Time: time.Now()}
	address := c.ResolvedIP
	if address == "" {
		address = c.Target
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(address, strconv.Itoa(c.Port)))
	s.RTT = float64(time.Since(s.Time)) / float64(time.Millisecond)
	if err != nil {
		s.Error = tcpError(err)
		if s.Error == "Conn_Timeout" || s.Error == "No_Route" || s.Error == "Network_Unreachable" {
			s.RTT = 0
		}
		return s
	}
	conn.Close()
	s.Success = true
	return s
}

// Keep the legacy TCP failure labels in live results and CSV exports.
func tcpError(err error) string {
	for _, item := range [][2]string{
		{"refused", "Conn_Refused"}, {"no route", "No_Route"},
		{"timeout", "Conn_Timeout"}, {"unreachable", "Network_Unreachable"},
	} {
		if strings.Contains(err.Error(), item[0]) {
			return item[1]
		}
	}
	return err.Error()
}
