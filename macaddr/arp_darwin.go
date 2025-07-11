//go:build darwin

package macaddr

// macOS implementation - will only be compiled on macOS/Darwin systems
func init() {
	// Override the default ARP table loader with the macOS-specific one
	darwinARPLoader = loadDarwinARPTable
}

// loadDarwinARPTable loads all entries from the macOS ARP table into the cache
func loadDarwinARPTable(r *Resolver) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if the table has already been loaded by another goroutine
	if r.arpLoaded {
		return
	}

	// macOS implementation would use syscalls to:
	// - call sysctl with CTL_NET, PF_ROUTE, 0, AF_INET, NET_RT_FLAGS, RTF_LLINFO
	// - parse the sockaddr structures to extract IP and MAC addresses

	// For now, just mark as loaded - placeholder for future implementation
	r.arpLoaded = true
}
