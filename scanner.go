package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"neti/macaddr"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// HostInfo represents information about a discovered host
type HostInfo struct {
	IP               string
	MAC              string
	Hostname         string
	ProcessTime      time.Duration // Total processing time (DNS, MAC, etc.)
	ICMPResponseTime time.Duration // ICMP ping response time
	OpenPorts        []int         // Discovered open ports
}

// ScanResult represents the result of scanning a subnet
type ScanResult struct {
	ReachableHosts []HostInfo
	Total          int
	Completed      int
}

// ProgressCallback is called during scanning to report progress
type ProgressCallback func(completed, total, found int)

// Scanner handles network scanning operations
type Scanner struct {
	Concurrency int
	Timeout     time.Duration
	macResolver *macaddr.Resolver
	UseTCP      bool
	UseUDP      bool
}

// NewScanner creates a new scanner with default settings
func NewScanner() *Scanner {
	return &Scanner{
		Concurrency: 20,
		Timeout:     500 * time.Millisecond,
		macResolver: macaddr.NewResolver(),
	}
}

// GetIPsFromSubnet converts a CIDR subnet to a list of IP addresses
func (s *Scanner) GetIPsFromSubnet(subnet string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast addresses for /24 and smaller subnets
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

// ScanSubnet scans a list of IPs and returns reachable ones with MAC addresses
func (s *Scanner) ScanSubnet(ips []string, progressCallback ProgressCallback) *ScanResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var reachableHosts []HostInfo
	var completed int

	semaphore := make(chan struct{}, s.Concurrency)
	total := len(ips)

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now() // Start timing for total process

			// First, try ICMP ping and measure its response time
			icmpReachable := false
			var icmpResponseTime time.Duration
			if reachable, responseTime := s.pingIP(ip); reachable {
				icmpReachable = true
				icmpResponseTime = responseTime
			}
			var openPorts []int
			// Separate TCP and UDP scanning so UDP probes are only run when the host is known
			// to be responsive (ICMP reply) or TCP scan found something. This avoids marking
			// many UDP ports as open|filtered for hosts that are likely down/unreachable.
			var tcpPorts []int
			var udpPorts []int

			if s.UseTCP {
				tcpPorts = s.getOpenPorts(ip)
			}

			if s.UseUDP {
				if icmpReachable || len(tcpPorts) > 0 {
					// Only perform UDP probes when host shows some responsiveness
					udpPorts = s.getOpenUDPPorts(ip)
				}
			}

			openPorts = append(openPorts, tcpPorts...)
			openPorts = append(openPorts, udpPorts...)

			// Host is considered reachable if found via ICMP or has open TCP ports
			isReachable := icmpReachable || len(openPorts) > 0

			if isReachable {
				var mac, hostname string

				// Only get MAC and hostname for ICMP-reachable hosts
				if icmpReachable {
					mac = s.macResolver.GetMACAddress(ip)

					// Perform reverse DNS lookup
					names, err := net.LookupAddr(ip)
					if err == nil && len(names) > 0 {
						// Return the first name, removing the trailing dot.
						hostname = strings.TrimSuffix(names[0], ".")
					}
				}
				// For TCP-only hosts, leave MAC and hostname empty

				processTime := time.Since(start) // Calculate duration

				mu.Lock()
				reachableHosts = append(reachableHosts, HostInfo{
					IP:               ip,
					MAC:              mac,
					Hostname:         hostname,
					ProcessTime:      processTime,
					ICMPResponseTime: icmpResponseTime,
					OpenPorts:        openPorts,
				})
				mu.Unlock()
			}

			// Update progress
			mu.Lock()
			completed++
			if progressCallback != nil {
				progressCallback(completed, total, len(reachableHosts))
			}
			mu.Unlock()
		}(ip)
	}

	wg.Wait()

	// Sort results for consistent output
	sort.Slice(reachableHosts, func(i, j int) bool {
		ip1 := net.ParseIP(reachableHosts[i].IP)
		ip2 := net.ParseIP(reachableHosts[j].IP)
		if ip1 != nil && ip2 != nil {
			ip1v4 := ip1.To4()
			ip2v4 := ip2.To4()
			if ip1v4 != nil && ip2v4 != nil {
				return binary.BigEndian.Uint32(ip1v4) < binary.BigEndian.Uint32(ip2v4)
			}
		}
		return reachableHosts[i].IP < reachableHosts[j].IP
	})

	return &ScanResult{
		ReachableHosts: reachableHosts,
		Total:          total,
		Completed:      completed,
	}
}

// getOpenPorts scans for open TCP ports on the target IP
func (s *Scanner) getOpenPorts(ip string) []int {
	commonPorts := []int{80, 443, 22, 21, 23, 25, 53, 135, 139, 445}
	var openPorts []int

	for _, port := range commonPorts {
		address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", address, s.Timeout)
		if err == nil {
			conn.Close()
			openPorts = append(openPorts, port)
		}
	}

	return openPorts
}

func (s *Scanner) getOpenUDPPorts(ip string) []int {
	udpPorts := []int{53, 67, 68, 69, 123, 137, 138, 161, 500, 514}
	var open []int
	dstIP := net.ParseIP(ip)

	for _, port := range udpPorts {
		raddr := &net.UDPAddr{IP: dstIP, Port: port}

		conn, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			// Can't dial UDP to this port — skip it
			continue
		}

		// Send a small probe. If the service replies on the same UDP socket we
		// consider the port open. Otherwise we treat it as closed/filtered and
		// do not report it.
		_ = conn.SetDeadline(time.Now().Add(s.Timeout))
		_, err = conn.Write([]byte("probe"))
		if err != nil {
			// Retry once on write error
			_ = conn.SetDeadline(time.Now().Add(s.Timeout))
			_, _ = conn.Write([]byte("probe"))
		}

		// Attempt to read a reply from the service.
		buf := make([]byte, 1500)
		_ = conn.SetReadDeadline(time.Now().Add(s.Timeout))
		n, _, err := conn.ReadFrom(buf)
		conn.Close()

		if err == nil && n > 0 {
			// Received application-layer response — consider port open.
			open = append(open, port)
		}
		// If no reply or read error, do not mark the port as open (avoid false positives).
	}

	return open
}

// pingIP sends an ICMP ping to an IP address and returns (success, duration)
func (s *Scanner) pingIP(ip string) (bool, time.Duration) {
	dst, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return false, 0
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false, 0
	}
	defer conn.Close()

	message := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("ping"),
		},
	}

	data, err := message.Marshal(nil)
	if err != nil {
		return false, 0
	}

	deadline := time.Now().Add(s.Timeout)
	conn.SetDeadline(deadline)

	start := time.Now()
	_, err = conn.WriteTo(data, dst)
	if err != nil {
		return false, 0
	}

	reply := make([]byte, 1500)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, peer, err := conn.ReadFrom(reply)
		if err != nil {
			continue
		}

		if peerIP, ok := peer.(*net.IPAddr); ok {
			if peerIP.IP.Equal(dst.IP) && len(reply) > 0 {
				return true, time.Since(start)
			}
		}
	}

	return false, 0
}

// incrementIP increments an IP address by one
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
