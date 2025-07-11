//go:build windows

package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows implementation - will only be compiled on Windows
func init() {
	// Override the default ARP table loader with the Windows-specific one
	windowsARPLoader = loadWindowsARPTable
}

// Windows API constants
const (
	NO_ERROR                  = 0
	ERROR_INSUFFICIENT_BUFFER = 122
)

// MIB_IPNETROW structure for GetIpNetTable function
type MIB_IPNETROW struct {
	Index       uint32
	PhysAddrLen uint32
	PhysAddr    [8]byte
	Addr        uint32
	Type        uint32
}

// loadWindowsARPTable loads all entries from the Windows ARP table into the cache
func loadWindowsARPTable(m *MACResolver) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if the table has already been loaded by another goroutine
	if m.arpLoaded {
		return
	}

	// Load required DLL and functions
	iphlpapi, err := windows.LoadDLL("iphlpapi.dll")
	if err != nil {
		return
	}
	defer iphlpapi.Release()

	getIpNetTableProc, err := iphlpapi.FindProc("GetIpNetTable")
	if err != nil {
		return
	}

	// First call to get required size
	var size uint32
	ret, _, _ := getIpNetTableProc.Call(0, uintptr(unsafe.Pointer(&size)), 1)

	if ret != ERROR_INSUFFICIENT_BUFFER {
		return
	}

	// Allocate buffer
	buf := make([]byte, size)
	ret, _, _ = getIpNetTableProc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0)

	if ret != NO_ERROR {
		return
	}

	// Parse table entries
	// First 4 bytes contain the number of entries
	count := *(*uint32)(unsafe.Pointer(&buf[0]))

	// Parse each entry
	for i := 0; i < int(count); i++ {
		// Calculate offset for this row
		rowSize := unsafe.Sizeof(MIB_IPNETROW{})
		offset := 4 + i*int(rowSize) // 4 bytes for the count, then the array of rows

		if offset+int(rowSize) > len(buf) {
			break
		}

		row := (*MIB_IPNETROW)(unsafe.Pointer(&buf[offset]))

		// Convert IP address from host byte order to network byte order
		ipBytes := make([]byte, 4)
		ipBytes[0] = byte(row.Addr)
		ipBytes[1] = byte(row.Addr >> 8)
		ipBytes[2] = byte(row.Addr >> 16)
		ipBytes[3] = byte(row.Addr >> 24)

		ipStr := fmt.Sprintf("%d.%d.%d.%d", ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])

		// Extract MAC address
		if row.PhysAddrLen == 6 {
			mac := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
				row.PhysAddr[0], row.PhysAddr[1], row.PhysAddr[2],
				row.PhysAddr[3], row.PhysAddr[4], row.PhysAddr[5])

			// Only store valid MACs
			if mac != "00:00:00:00:00:00" {
				m.cache[ipStr] = mac
			}
		}
	}

	m.arpLoaded = true
}
