package main

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/schollz/progressbar/v3"
	"os"
)

// UI handles user interface operations
type UI struct {
	progressBar *progressbar.ProgressBar
}

// NewUI creates a new UI instance
func NewUI() *UI {
	return &UI{}
}

// ShowUsage displays usage information
func (ui *UI) ShowUsage(programName string) {
	fmt.Printf("Usage: %s <subnet>\n", programName)
	fmt.Println("Example: go run main.go 192.168.1.0/24")
}

// ShowError displays an error message
func (ui *UI) ShowError(message string, err error) {
	fmt.Printf("%s: %v\n", message, err)
}

// ShowScanStart displays scan initialization information
func (ui *UI) ShowScanStart(subnet string, totalIPs int) {
	fmt.Printf("Scanning subnet: %s\n", subnet)
	fmt.Printf("Found %d IPs to scan\n", totalIPs)
	fmt.Printf("Scanning %d IPs...", totalIPs)
	ui.progressBar = progressbar.NewOptions(totalIPs,
		progressbar.OptionSetDescription("Scanning"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionClearOnFinish(),
	)
}

// ShowProgress displays scanning progress
func (ui *UI) ShowProgress(completed, total, found int) {
	if ui.progressBar != nil {
		ui.progressBar.Set(completed)
	}
}

// ShowResults displays the final scan results.
func (ui *UI) ShowResults(result *ScanResult) {
	fmt.Println() // New line after progress

	if len(result.ReachableHosts) == 0 {
		fmt.Println("\nNo reachable hosts found.")
		fmt.Println("Scan complete.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredDark)
	t.AppendHeader(table.Row{"#", "IP Address", "Hostname", "MAC Address", "Manufacturer"})

	for i, host := range result.ReachableHosts {
		mac := host.MAC
		vendor := mac2manufacturer(mac)
		t.AppendRow(table.Row{i + 1, host.IP, host.Hostname, mac, vendor})
	}

	t.Render()
	fmt.Printf("Scan complete. (%d/%d hosts responded)\n", len(result.ReachableHosts), result.Total)
}
