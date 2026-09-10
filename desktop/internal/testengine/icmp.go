package testengine

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func validateICMP(c Config) (Config, error) {
	// Hostname and timer constraints are shared with connection tests.
	target := c
	target.Port = 1
	checked, err := validateTCP(target)
	if err != nil {
		return c, errors.New(strings.ReplaceAll(err.Error(), "TCP", "ICMP"))
	}
	if c.PayloadSize < 32 || c.PayloadSize > 65507 {
		return c, errors.New("ICMP payload must be between 32 and 65507 bytes (IPv4 packet limit)")
	}
	if ip := net.ParseIP(checked.Target); ip != nil && ip.To4() == nil {
		return c, errors.New("ICMP tests require an IPv4 address or hostname with an IPv4 address")
	}
	if checked.ResolvedIP != "" && net.ParseIP(checked.ResolvedIP).To4() == nil {
		return c, errors.New("ICMP resolved address must be IPv4")
	}
	return Config{Type: "icmp", Target: checked.Target, URL: checked.Target, ResolvedIP: checked.ResolvedIP,
		PayloadSize: c.PayloadSize, DF: c.DF, Recording: c.Recording, IntervalMS: c.IntervalMS, TimeoutMS: c.TimeoutMS, AcceptedStatuses: []string{}}, nil
}

func resolveICMP(c Config) (Config, error) {
	if c.ResolvedIP != "" {
		return c, nil
	}
	if ip := net.ParseIP(c.Target).To4(); ip != nil {
		c.ResolvedIP = ip.String()
		return c, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), min(10*time.Second, time.Duration(c.TimeoutMS)*time.Millisecond))
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip4", c.Target)
	if err != nil {
		return c, fmt.Errorf("Cannot resolve ICMP target %s: %w", c.Target, err)
	}
	if len(addresses) == 0 {
		return c, fmt.Errorf("No IPv4 address for ICMP target %s", c.Target)
	}
	c.ResolvedIP = addresses[0].String()
	return c, nil
}

// Resolve the entire batch before creating sessions or sending echo requests.
func (r *Runner) StartICMP(config Config, targets string) ([]Session, error) {
	if len(targets) > 65536 {
		return nil, errors.New("ICMP target list is too long")
	}
	var configs []Config
	for _, line := range strings.Split(targets, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		config.Type, config.Target, config.ResolvedIP = "icmp", strings.TrimSpace(line), ""
		c, err := validateICMP(config)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
		if len(configs) > MaxActive {
			return nil, errors.New("8 active tests maximum")
		}
	}
	if len(configs) == 0 {
		return nil, errors.New("Enter at least one ICMP server IP or hostname")
	}
	for i, c := range configs {
		resolved, err := resolveICMP(c)
		if err != nil {
			return nil, err
		}
		configs[i] = resolved
	}
	return r.startBatch(configs)
}

func probeICMP(ctx context.Context, c Config) Sample {
	if err := ctx.Err(); err != nil {
		return Sample{Time: time.Now(), Error: err.Error()}
	}
	if !c.DF && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
		conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
		if err == nil {
			return probeICMPSocket(ctx, c, conn)
		}
	}
	return probeICMPCommand(ctx, c)
}

// A separate datagram socket and random echoed payload correlate each probe.
// Closing the socket on cancellation makes Stop and application exit immediate.
func probeICMPSocket(ctx context.Context, c Config, conn *icmp.PacketConn) Sample {
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	s := Sample{Time: time.Now()}
	payload := make([]byte, c.PayloadSize)
	if _, err := rand.Read(payload); err != nil {
		s.Error = err.Error()
		return s
	}
	request := icmp.Message{Type: ipv4.ICMPTypeEcho, Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: payload}}
	wire, err := request.Marshal(nil)
	if err != nil {
		s.Error = err.Error()
		return s
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(time.Duration(c.TimeoutMS) * time.Millisecond)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		s.Error = err.Error()
		return s
	}
	s.Time = time.Now()
	if _, err := conn.WriteTo(wire, &net.UDPAddr{IP: net.ParseIP(c.ResolvedIP)}); err != nil {
		s.Error = err.Error()
		return s
	}
	buffer := make([]byte, 65535)
	for {
		n, peer, err := conn.ReadFrom(buffer)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				s.Error = "Timeout"
			} else {
				s.Error = err.Error()
			}
			return s
		}
		if matchesICMPEcho(buffer[:n], peer, c.ResolvedIP, payload) {
			s.Success, s.RTT = true, float64(time.Since(s.Time))/float64(time.Millisecond)
			return s
		}
		// Ignore unrelated ICMP traffic. An unanswered request expires at its deadline.
	}
}

func matchesICMPEcho(wire []byte, peer net.Addr, address string, payload []byte) bool {
	ip := net.ParseIP(address)
	var peerIP net.IP
	switch p := peer.(type) {
	case *net.UDPAddr:
		peerIP = p.IP
	case *net.IPAddr:
		peerIP = p.IP
	}
	if !peerIP.Equal(ip) {
		return false
	}
	reply, err := icmp.ParseMessage(1, wire)
	if err != nil || reply.Type != ipv4.ICMPTypeEchoReply || reply.Code != 0 {
		return false
	}
	echo, ok := reply.Body.(*icmp.Echo)
	return ok && echo.Seq == 1 && bytes.Equal(echo.Data, payload)
}

func icmpCommandArgs(c Config, platform string) ([]string, error) {
	size, timeout := strconv.Itoa(c.PayloadSize), strconv.Itoa(c.TimeoutMS)
	var args []string
	switch platform {
	case "darwin":
		args = []string{"-n", "-c", "1", "-W", timeout, "-s", size}
		if c.DF {
			args = append(args, "-D")
		}
	case "linux":
		args = []string{"-n", "-c", "1", "-W", strconv.Itoa(c.TimeoutMS / 1000), "-s", size}
		if c.DF {
			args = append(args, "-M", "do")
		}
	case "windows":
		args = []string{"-4", "-n", "1", "-w", timeout, "-l", size}
		if c.DF {
			args = append(args, "-f")
		}
	default:
		return nil, errors.New("ICMP is supported on macOS, Linux, and Windows")
	}
	return append(args, c.ResolvedIP), nil
}

// Retain a bounded diagnostic even if an installed ping command is unusually noisy.
type pingOutput struct{ bytes.Buffer }

func (b *pingOutput) Write(p []byte) (int, error) {
	n := len(p)
	if left := 8192 - b.Len(); left > 0 {
		_, _ = b.Buffer.Write(p[:min(left, n)])
	}
	return n, nil
}

func probeICMPCommand(ctx context.Context, c Config) Sample {
	s := Sample{Time: time.Now()}
	args, err := icmpCommandArgs(c, runtime.GOOS)
	if err != nil {
		s.Error = err.Error()
		return s
	}
	cmd := exec.CommandContext(ctx, "ping", args...)
	configurePingCommand(cmd)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	cmd.WaitDelay = 100 * time.Millisecond
	var output pingOutput
	cmd.Stdout, cmd.Stderr = &output, &output
	err = cmd.Run()
	s.Success, s.RTT, s.Error = parseICMPOutput(output.String())
	if ctx.Err() != nil {
		s.Success, s.RTT, s.Error = false, 0, "Timeout"
		return s
	}
	if err != nil && s.Error == "Unknown_Error" {
		s.Error = fmt.Sprintf("ICMP ping failed: %v: %s", err, strings.TrimSpace(output.String()))
	}
	return s
}

var icmpRTT = regexp.MustCompile(`(?i)time\s*[=<]\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)

func parseICMPOutput(output string) (bool, float64, string) {
	lower := strings.ToLower(output)
	for _, item := range [][2]string{{"message too long", "MTU Exceed, DF set"}, {"frag needed", "MTU Exceed, DF set"}, {"needs to be fragmented", "MTU Exceed, DF set"}, {"unreachable", "NoRoute"}, {"no route", "NoRoute"}, {"time to live exceeded", "TTL_Exceeded"}, {"ttl expired", "TTL_Exceeded"}} {
		if strings.Contains(lower, item[0]) {
			return false, 0, item[1]
		}
	}
	if match := icmpRTT.FindStringSubmatch(output); len(match) == 2 && (strings.Contains(lower, "bytes from") || strings.Contains(lower, "ttl=")) {
		rtt, err := strconv.ParseFloat(match[1], 64)
		if err == nil {
			return true, rtt, ""
		}
	}
	if strings.Contains(lower, "100% packet loss") || strings.Contains(lower, "100.0% packet loss") || strings.Contains(lower, "timed out") || strings.Contains(lower, "timeout") {
		return false, 0, "Timeout"
	}
	return false, 0, "Unknown_Error"
}
