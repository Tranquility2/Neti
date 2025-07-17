package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const ouiFileURL = "http://standards-oui.ieee.org/oui/oui.txt"
const ouiFileName = "oui.txt"

// OUI cache
var (
	ouiCache         map[string]string
	loadOUICacheOnce sync.Once
)

// updateOUIFile fetches the OUI file from the IEEE website and saves it locally.
func updateOUIFile() error {
	if _, err := os.Stat(ouiFileName); err == nil {
		fmt.Printf("(OUI file already exists, skipping download.)")
		return nil
	}

	fmt.Printf("\n(Downloading OUI file from IEEE...)")

	resp, err := http.Get(ouiFileURL)
	if err != nil {
		return fmt.Errorf("failed to download OUI file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download OUI file: received status code %d", resp.StatusCode)
	}

	file, err := os.Create(ouiFileName)
	if err != nil {
		return fmt.Errorf("failed to create OUI file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save OUI file: %w", err)
	}

	return nil
}

// loadOUICache loads the OUI file into an in-memory map.
func loadOUICache() {
	ouiCache = make(map[string]string)
	file, err := os.Open(ouiFileName)
	if err != nil {
		// If the file doesn't exist, the cache will simply be empty.
		// Lookups will fail gracefully.
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "(base 16)") {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}

		// The OUI prefix is part of the first segment, e.g., "00-00-01   (base 16)"
		ouiPrefix := strings.Fields(parts[0])[0]
		ouiPrefix = strings.ReplaceAll(ouiPrefix, "-", "")

		// The vendor is the second part, trimmed of whitespace.
		vendor := strings.TrimSpace(parts[1])
		ouiCache[ouiPrefix] = vendor
	}
}

// mac2manufacturer looks up the manufacturer for a given MAC address from the in-memory OUI cache.
func mac2manufacturer(mac string) string {
	// Ensure the OUI cache is loaded, but only once.
	loadOUICacheOnce.Do(loadOUICache)

	// Normalize MAC to OUI prefix (e.g., 00:1A:2B:3C:4D:5E -> 001A2B)
	macPrefix := strings.ToUpper(strings.ReplaceAll(mac, ":", ""))
	if len(macPrefix) < 6 {
		return "Invalid MAC"
	}
	macPrefix = macPrefix[:6]

	if vendor, ok := ouiCache[macPrefix]; ok {
		return vendor
	}

	return ""
}
