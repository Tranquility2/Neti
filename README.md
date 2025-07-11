# Neti

A fast, cross-platform network scanner that discovers hosts, MAC addresses, hostnames, and hardware vendors.

## Features

- **Concurrent ICMP Scanning**: Quickly scans a subnet using goroutines.
- **Cross-Platform**: Natively supports Linux, Windows, and macOS.
- **MAC Address Resolution**: Fetches MAC addresses using platform-specific APIs.
- **Hostname Resolution**: Performs reverse DNS lookups to find hostnames.
- **OUI Vendor Lookup**: Identifies the hardware manufacturer from the MAC address.
- **Clean Table Output**: Displays results in a well-aligned, easy-to-read table.
- **Progress Indicator**: Shows real-time scan progress.

## Example Output

```
Scanning subnet: 192.168.1.0/24
Found 254 IPs to scan
Scanning 254 IPs...
Progress: 254/254 (100.0%) - Found 3 hosts

Found 3 reachable hosts:
┌─────┬─────────────────┬─────────────────────────┬───────────────────┬──────────────────────────────────────────┐
│  #  │ IP Address      │ Hostname                │ MAC Address       │ Manufacturer                             │
├─────┼─────────────────┼─────────────────────────┼───────────────────┼──────────────────────────────────────────┤
│   1 │ 192.168.1.1     │ router.local            │ 1A:2B:3C:4D:5E:6F │ NETGEAR                                  │
│   2 │ 192.168.1.10    │ my-laptop               │ 9F:8E:7D:6C:5B:4A │ Apple, Inc.                              │
│   3 │ 192.168.1.50    │ N/A                     │ 11:22:33:44:55:66 │ Intel Corporate                          │
└─────┴─────────────────┴─────────────────────────┴───────────────────┴──────────────────────────────────────────┘
Scan complete. (3/254 hosts responded)
```

## Setup & Usage

**1. Install Dependencies**

This will fetch the necessary Go packages.

```bash
make deps
```

**2. Download the OUI Vendor File**

For manufacturer lookups, you need the OUI file from the IEEE. This only needs to be done once.

```bash
# This command is not yet in the Makefile, run this manually for now:
go run . update-oui 
# Or, if you have an older version of the code, you can add an `update-oui` command to your Makefile.
```
*Note: A future version should integrate this into the Makefile.*

**3. Run a Scan**

You must run the scanner with `sudo` because it requires raw socket permissions to send ICMP packets.

```bash
make run-sudo SUBNET=192.168.1.0/24
```

## Building

You can build the binary for your current operating system or for all supported platforms.

```bash
# Build for your current OS
make build

# Build for Linux, Windows, and macOS
make build-all
```
The binaries will be placed in the `build/` directory.
