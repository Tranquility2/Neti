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
	IP          string
	MAC         string
	Hostname    string
	ProcessTime time.Duration
	OpenPorts   []int         // New field for discovered open ports
	FoundVia    string        // "ICMP", "TCP", or "ICMP+TCP"
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

			start := time.Now() // Start timing

			// First, try ICMP ping
			icmpReachable := s.pingIP(ip)
			var openPorts []int
			var foundVia string

			if icmpReachable {
				foundVia = "ICMP"
				// If TCP scanning is enabled, also check for open ports
				if s.UseTCP {
					openPorts = s.getOpenPorts(ip)
					if len(openPorts) > 0 {
						foundVia = "ICMP+TCP"
					}
				}
			} else if s.UseTCP {
				// If ICMP failed but TCP is enabled, try TCP-only discovery
				openPorts = s.getOpenPorts(ip)
				if len(openPorts) > 0 {
					foundVia = "TCP"
				}
			}

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
					IP:          ip,
					MAC:         mac,
					Hostname:    hostname,
					ProcessTime: processTime,
					OpenPorts:   openPorts,
					FoundVia:    foundVia,
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

// tcpConnect attempts to connect to common ports on the target IP
func (s *Scanner) tcpConnect(ip string) bool {
	// Try a few very common ports
	commonPorts := []int{80, 443, 22}
	
	var hasConnectionRefused bool
	
	for _, port := range commonPorts {
		address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", address, s.Timeout)
		if err == nil {
			// Successfully connected - host is definitely up
			conn.Close()
			return true
		}
		
		// Check the type of error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// This was a timeout - continue to next port
			continue
		}
		
		// If we get here, it's likely a "connection refused" error
		// which means the host is up but the port is closed
		hasConnectionRefused = true
	}
	
	// If we got connection refused on any port, the host is likely up
	return hasConnectionRefused
}

// pingIP sends an ICMP ping to an IP address
func (s *Scanner) pingIP(ip string) bool {
	dst, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return false
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false
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
		return false
	}

	deadline := time.Now().Add(s.Timeout)
	conn.SetDeadline(deadline)

	_, err = conn.WriteTo(data, dst)
	if err != nil {
		return false
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
				return true
			}
		}
	}

	return false
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
