package main

import "os"

func main() {
	ui := NewUI()
	scanner := NewScanner()

	if len(os.Args) != 2 {
		ui.ShowUsage(os.Args[0])
		os.Exit(1)
	}

	subnet := os.Args[1]

	ips, err := scanner.GetIPsFromSubnet(subnet)
	if err != nil {
		ui.ShowError("Error parsing subnet", err)
		os.Exit(1)
	}

	ui.ShowScanStart(subnet, len(ips))

	result := scanner.ScanSubnet(ips, ui.ShowProgress)

	ui.ShowResults(result)
}
