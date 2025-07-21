package main

import (
	"flag"
	"os"
	"path/filepath"
)

func main() {
	ui := NewUI()
	scanner := NewScanner()

	var subnet string
	flag.StringVar(&subnet, "subnet", "", "CIDR subnet to scan (e.g. 192.168.1.0/24)")
	flag.Parse()

	// Support positional argument as subnet
	if subnet == "" && flag.NArg() > 0 {
		subnet = flag.Arg(0)
	}

	if subnet == "" {
		ui.ShowUsage(filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	ips, err := scanner.GetIPsFromSubnet(subnet)
	if err != nil {
		ui.ShowError("Error parsing subnet", err)
		os.Exit(1)
	}

	ui.ShowScanStart(subnet, len(ips))

	result := scanner.ScanSubnet(ips, ui.ShowProgress)

	updateOUIFile()

	ui.ShowResults(result)
}
