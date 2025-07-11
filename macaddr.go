package main

import (
	"net"
	"runtime"
	"strings"
	"sync"
)

// MACResolver handles MAC address resolution for different platforms
type MACResolver struct {
	// Cache of IP to MAC mappings to avoid repeated lookups
	cache map[string]string
	// Flag to indicate if ARP table has been loaded
	arpLoaded bool
	// Platform-specific ARP table loader function
	loadARPTableFunc func(*MACResolver)
	// Mutex to protect concurrent access to the cache and arpLoaded flag
	mutex sync.Mutex
}

// Platform-specific function type for loading ARP tables
type ARPTableLoader func(*MACResolver)

// Platform-specific loaders that will be set in init()
var (
	linuxARPLoader   ARPTableLoader
	windowsARPLoader ARPTableLoader
	darwinARPLoader  ARPTableLoader
)

// init sets up the appropriate platform-specific functions
func init() {
	// These will be overridden in platform-specific files
}

// NewMACResolver creates a new MAC resolver
func NewMACResolver() *MACResolver {
	resolver := &MACResolver{
		cache:     make(map[string]string),
		arpLoaded: false,
	}

	// Set the appropriate ARP table loader based on platform
	switch runtime.GOOS {
	case "linux":
		resolver.loadARPTableFunc = linuxARPLoader
	case "windows":
		resolver.loadARPTableFunc = windowsARPLoader
	case "darwin":
		resolver.loadARPTableFunc = darwinARPLoader
	}

	// Pre-load the ARP table if a loader is available
	if resolver.loadARPTableFunc != nil {
		resolver.loadARPTableFunc(resolver)
	} else {
		// For unsupported platforms, just mark as loaded
		resolver.arpLoaded = true
	}

	return resolver
}

// GetMACAddress gets the MAC address for an IP using platform-specific methods
func (m *MACResolver) GetMACAddress(ip string) string {
	// First check the cache for previously resolved MAC addresses
	if mac := m.getMACFromCache(ip); mac != "" {
		return mac
	}

	// getMACFromLocalInterfaces can be slow, so we run it outside the lock.
	// It doesn't access shared state.
	if mac := m.getMACFromLocalInterfaces(ip); mac != "" {
		// Lock before writing to the shared cache
		m.mutex.Lock()
		m.cache[ip] = mac
		m.mutex.Unlock()
		return mac
	}

	// If not a local IP, try platform-specific ARP table lookups
	m.ensureARPTableLoaded()

	// Check cache again after platform-specific ARP table load
	if mac := m.getMACFromCache(ip); mac != "" {
		return mac
	}

	// Try reloading the ARP table - it might have been updated
	m.reloadARPTable()

	// Final cache check
	return m.getMACFromCache(ip)
}

// getMACFromCache checks if an IP address is in the cache
func (m *MACResolver) getMACFromCache(ip string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if mac, ok := m.cache[ip]; ok {
		return mac
	}
	return ""
}

// ensureARPTableLoaded makes sure the ARP table is loaded for the current platform
func (m *MACResolver) ensureARPTableLoaded() {
	m.mutex.Lock()
	// Check if already loaded while holding the lock
	if m.arpLoaded {
		m.mutex.Unlock()
		return
	}
	m.mutex.Unlock() // Unlock before calling the loader function

	// Use the platform-specific loader function if available
	if m.loadARPTableFunc != nil {
		m.loadARPTableFunc(m)
	}
}

// reloadARPTable forces a reload of the platform-specific ARP table
func (m *MACResolver) reloadARPTable() {
	// Reset the flag and reload. The lock inside the loader will handle synchronization.
	if m.loadARPTableFunc != nil {
		m.loadARPTableFunc(m)
	}
}

// --- Helper Functions ---

// getMACFromLocalInterfaces checks if the IP belongs to a local network interface
func (m *MACResolver) getMACFromLocalInterfaces(ip string) string {
	targetIP := net.ParseIP(ip)
	if targetIP == nil {
		return ""
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.Equal(targetIP) {
					return strings.ToUpper(iface.HardwareAddr.String())
				}
			}
		}
	}
	return ""
}

func isValidMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil && mac != "00:00:00:00:00:00"
}
