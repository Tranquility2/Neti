package macaddr

import (
	"net"
	"runtime"
	"strings"
	"sync"
)

// Resolver handles MAC address resolution for different platforms.
type Resolver struct {
	// Cache of IP to MAC mappings to avoid repeated lookups
	cache map[string]string
	// Flag to indicate if ARP table has been loaded
	arpLoaded bool
	// Platform-specific ARP table loader function
	loadARPTableFunc func(*Resolver)
	// Mutex to protect concurrent access to the cache and arpLoaded flag
	mutex sync.Mutex
}

// ARPTableLoader is a platform-specific function type for loading ARP tables.
type ARPTableLoader func(*Resolver)

// Platform-specific loaders that will be set in init()
var (
	linuxARPLoader   ARPTableLoader
	windowsARPLoader ARPTableLoader
	darwinARPLoader  ARPTableLoader
)

// NewResolver creates a new MAC address resolver.
func NewResolver() *Resolver {
	resolver := &Resolver{
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

// GetMACAddress gets the MAC address for an IP using platform-specific methods.
func (r *Resolver) GetMACAddress(ip string) string {
	// First check the cache for previously resolved MAC addresses
	if mac := r.getMACFromCache(ip); mac != "" {
		return mac
	}

	// getMACFromLocalInterfaces can be slow, so we run it outside the lock.
	// It doesn't access shared state.
	if mac := r.getMACFromLocalInterfaces(ip); mac != "" {
		// Lock before writing to the shared cache
		r.mutex.Lock()
		r.cache[ip] = mac
		r.mutex.Unlock()
		return mac
	}

	// If not a local IP, try platform-specific ARP table lookups
	r.ensureARPTableLoaded()

	// Check cache again after platform-specific ARP table load
	if mac := r.getMACFromCache(ip); mac != "" {
		return mac
	}

	// Try reloading the ARP table - it might have been updated
	r.reloadARPTable()

	// Final cache check
	mac := r.getMACFromCache(ip)
	if mac != "" {
		return mac
	}

	// Fallback: send ARP request and reload ARP table
	sendARPRequest(ip)
	r.reloadARPTable()
	return r.getMACFromCache(ip)
}

// getMACFromCache checks if an IP address is in the cache.
func (r *Resolver) getMACFromCache(ip string) string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if mac, ok := r.cache[ip]; ok {
		return mac
	}
	return ""
}

// sendARPRequest sends a dummy UDP packet to the target IP to trigger an ARP request.
func sendARPRequest(ip string) {
	conn, err := net.Dial("udp", ip+":0")
	if err == nil {
		conn.Close()
	}
}

// ensureARPTableLoaded makes sure the ARP table is loaded for the current platform.
func (r *Resolver) ensureARPTableLoaded() {
	r.mutex.Lock()
	// Check if already loaded while holding the lock
	if r.arpLoaded {
		r.mutex.Unlock()
		return
	}
	r.mutex.Unlock() // Unlock before calling the loader function

	// Use the platform-specific loader function if available
	if r.loadARPTableFunc != nil {
		r.loadARPTableFunc(r)
	}
}

// reloadARPTable forces a reload of the platform-specific ARP table.
func (r *Resolver) reloadARPTable() {
	// Reset the flag and reload. The lock inside the loader will handle synchronization.
	if r.loadARPTableFunc != nil {
		r.loadARPTableFunc(r)
	}
}

// getMACFromLocalInterfaces checks if the IP belongs to a local network interface.
func (r *Resolver) getMACFromLocalInterfaces(ip string) string {
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
