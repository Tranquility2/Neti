//go:build linux
// +build linux

package main

import (
	"bufio"
	"os"
	"strings"
)

// Linux implementation - will only be compiled on Linux
func init() {
	// Override the default ARP table loader with the Linux-specific one
	linuxARPLoader = loadLinuxARPTable
}

// loadLinuxARPTable is the Linux-specific implementation for loading the ARP table
func loadLinuxARPTable(m *MACResolver) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if the table has already been loaded by another goroutine
	if m.arpLoaded {
		return
	}

	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // Skip header

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			ip := fields[0]
			mac := fields[3]
			if isValidMAC(mac) {
				m.cache[ip] = strings.ToUpper(mac)
			}
		}
	}

	m.arpLoaded = true
}
