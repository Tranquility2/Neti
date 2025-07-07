package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <subnet>\n", os.Args[0])
		fmt.Println("Example: go run main.go 192.168.1.0/24")
		os.Exit(1)
	}

	subnet := os.Args[1]
	fmt.Printf("Scanning subnet: %s\n", subnet)

	ips, err := getIPsFromSubnet(subnet)
	if err != nil {
		fmt.Printf("Error parsing subnet: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d IPs to scan\n", len(ips))

	scanSubnet(ips)
}

func getIPsFromSubnet(subnet string) ([]string, error) {
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

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func scanSubnet(ips []string) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var reachableIPs []string

	// Limit concurrent pings to avoid overwhelming the system
	semaphore := make(chan struct{}, 20)

	fmt.Printf("Scanning %d IPs...\n", len(ips))

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			if pingIP(ip) {
				mu.Lock()
				reachableIPs = append(reachableIPs, ip)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()

	// Sort results for consistent output
	sort.Slice(reachableIPs, func(i, j int) bool {
		// Parse IPs and compare numerically
		ip1 := net.ParseIP(reachableIPs[i])
		ip2 := net.ParseIP(reachableIPs[j])
		if ip1 != nil && ip2 != nil {
			ip1v4 := ip1.To4()
			ip2v4 := ip2.To4()
			if ip1v4 != nil && ip2v4 != nil {
				return binary.BigEndian.Uint32(ip1v4) < binary.BigEndian.Uint32(ip2v4)
			}
		}
		return reachableIPs[i] < reachableIPs[j]
	})

	// Print sorted results
	fmt.Printf("\nFound %d reachable hosts:\n", len(reachableIPs))
	for _, ip := range reachableIPs {
		fmt.Printf("✓ %s is reachable\n", ip)
	}

	fmt.Println("Scan complete.")
}

func pingIP(ip string) bool {
	// Parse destination IP first
	dst, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return false
	}

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return false
	}
	defer conn.Close()

	// Create ICMP Echo Request message with unique ID
	message := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("ping"),
		},
	}

	// Marshal the message
	data, err := message.Marshal(nil)
	if err != nil {
		return false
	}

	// Set timeout
	deadline := time.Now().Add(500 * time.Millisecond)
	conn.SetDeadline(deadline)

	// Send the packet
	_, err = conn.WriteTo(data, dst)
	if err != nil {
		return false
	}

	// Read the reply with timeout
	reply := make([]byte, 1500)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, peer, err := conn.ReadFrom(reply)
		if err != nil {
			continue // Timeout or other error, keep trying
		}

		// Check if response is from our target IP
		if peerIP, ok := peer.(*net.IPAddr); ok {
			if peerIP.IP.Equal(dst.IP) && len(reply) > 0 {
				return true
			}
		}
	}

	return false
}
