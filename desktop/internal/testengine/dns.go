package testengine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

func validateDNS(c Config) (Config, error) {
	c.Resolver = strings.TrimSpace(c.Resolver)
	if net.ParseIP(c.Resolver) == nil {
		return c, fmt.Errorf("DNS server must be a valid IP address. Invalid input: %s", c.Resolver)
	}
	if strings.TrimSpace(c.Query) == "" {
		return c, errors.New("DNS Query cannot be empty!")
	}
	if len(c.Query) > 4096 {
		return c, errors.New("DNS query is too long")
	}
	if c.Protocol == "" {
		c.Protocol = "udp"
	}
	if c.Protocol != "udp" && c.Protocol != "tcp" {
		return c, errors.New("Choose udp or tcp for DNS protocol")
	}
	const maxDurationMS = int64((1<<63 - 1) / time.Millisecond)
	if c.IntervalMS < 1000 || c.IntervalMS%1000 != 0 || int64(c.IntervalMS) > maxDurationMS {
		return c, errors.New("Interval must be a positive whole number of seconds")
	}
	if c.TimeoutMS < 1000 || c.TimeoutMS%1000 != 0 || int64(c.TimeoutMS) > maxDurationMS {
		return c, errors.New("Timeout must be a positive whole number of seconds")
	}
	// The existing search column also holds a human-readable DNS target. No
	// schema migration or alternate session architecture is required.
	return Config{Type: "dns", Resolver: c.Resolver, Query: c.Query, Protocol: c.Protocol,
		Recording: c.Recording, IntervalMS: c.IntervalMS, TimeoutMS: c.TimeoutMS,
		URL: c.Resolver + " · " + c.Query, AcceptedStatuses: []string{}}, nil
}

// StartDNS validates every resolver before starting any workers. A failed
// batch is rolled back while the runner lock still prevents probe delivery.
func (r *Runner) StartDNS(config Config, resolvers string) ([]Session, error) {
	if len(resolvers) > 65536 {
		return nil, errors.New("Resolver list is too long")
	}
	var configs []Config
	for _, line := range strings.Split(resolvers, "\n") {
		if line == "" || line == "\r" {
			continue
		}
		config.Type, config.Resolver = "dns", strings.TrimSpace(line)
		c, err := validateDNS(config)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	if len(configs) == 0 {
		return nil, errors.New("No Input IP/Host Target")
	}
	return r.startBatch(configs)
}

func newDNSResolver(protocol, address string) *net.Resolver {
	return &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, protocol, address)
	}}
}

func probeDNS(ctx context.Context, c Config) Sample {
	return probeDNSWithResolver(ctx, c.Query, newDNSResolver(c.Protocol, net.JoinHostPort(c.Resolver, "53")))
}

// Match nt v1.4.0 DnsProbing used by Fyne: LookupHost, IPv4-only response
// text, RTT before the CNAME lookup, and A/CNAME classification. The parent
// context additionally permits prompt cancellation on Stop and app exit.
func probeDNSWithResolver(ctx context.Context, query string, resolver *net.Resolver) Sample {
	s := Sample{Time: time.Now()}
	addresses, err := resolver.LookupHost(ctx, query)
	s.RTT = float64(time.Since(s.Time)) / float64(time.Millisecond)
	var ipv4 []string
	for _, address := range addresses {
		if ip := net.ParseIP(address); ip != nil && ip.To4() != nil {
			ipv4 = append(ipv4, address)
		}
	}
	s.DNSResponse = strings.Join(ipv4, ",")
	if err != nil {
		s.Error = dnsError(err)
		return s
	}
	cname, cnameErr := resolver.LookupCNAME(ctx, query)
	s.Success = true
	if cnameErr == nil {
		cname = strings.TrimSuffix(cname, ".")
	}
	s.DNSRecord = "A"
	if cname != query {
		s.DNSRecord = "CNAME"
	}
	return s
}

func dnsError(err error) string {
	for _, item := range [][2]string{
		{"no such host", "Non_Existent_Domain"}, {"timeout", "Timeout"},
		{"deadline exceeded", "Timeout"}, {"temporary failure", "Temporary_failure"},
		{"query refused", "DNS_query_refused"}, {"invalid domain name", "Invalid_domain_name"},
		{"permission denied", "Permission_denied"},
	} {
		if strings.Contains(err.Error(), item[0]) {
			return item[1]
		}
	}
	return "dns query failed: " + err.Error()
}
