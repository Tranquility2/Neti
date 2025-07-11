package main

import (
	"encoding/binary"
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
	IP       string
	MAC      string
	Hostname string
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

			if s.pingIP(ip) {
				mac := s.macResolver.GetMACAddress(ip)

				// Perform reverse DNS lookup
				var hostname string
				names, err := net.LookupAddr(ip)
				if err == nil && len(names) > 0 {
					// Return the first name, removing the trailing dot.
					hostname = strings.TrimSuffix(names[0], ".")
				} else {
					hostname = ""
				}

				mu.Lock()
				reachableHosts = append(reachableHosts, HostInfo{
					IP:       ip,
					MAC:      mac,
					Hostname: hostname,
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
