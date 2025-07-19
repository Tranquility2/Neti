package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
)

// UI handles user interface operations
type UI struct {
	progressWriter progress.Writer
	tracker        *progress.Tracker
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

	ui.tracker = &progress.Tracker{
		Message: "Scanning",
		Total:   int64(totalIPs),
		Units:   progress.UnitsDefault,
	}
	ui.progressWriter = progress.NewWriter()
	ui.progressWriter.SetOutputWriter(os.Stdout)
	ui.progressWriter.SetStyle(progress.StyleBlocks)
	ui.progressWriter.ShowETA(true)
	ui.progressWriter.AppendTracker(ui.tracker)
	go ui.progressWriter.Render()
}

// ShowProgress displays scanning progress
func (ui *UI) ShowProgress(completed, total, found int) {
	if ui.tracker != nil {
		ui.tracker.SetValue(int64(completed))
	}
}

func formatProcessTime(d time.Duration) string {
	ms := d.Milliseconds()
	if ms >= 1000 {
		return fmt.Sprintf("\033[31m%ds\033[0m", int(ms/1000)) // Red color for seconds
	} else if ms >= 50 {
		return fmt.Sprintf("\033[33m%dms\033[0m", ms) // Yellow color
	} else if ms <= 20 {
		return fmt.Sprintf("\033[32m%dms\033[0m", ms) // Green color
	}
	return fmt.Sprintf("%dms", ms)
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
	t.AppendHeader(table.Row{"#", "IP Address", "Hostname", "MAC Address", "Manufacturer", "Process Time"})

	for i, host := range result.ReachableHosts {
		mac := host.MAC
		vendor := mac2manufacturer(mac)
		processTimeStr := formatProcessTime(host.ProcessTime)
		t.AppendRow(table.Row{i + 1, host.IP, host.Hostname, mac, vendor, processTimeStr})
	}

	t.Render()
	fmt.Printf("Scan complete. (%d/%d hosts responded)\n", len(result.ReachableHosts), result.Total)
}
