# Neti

A fast network scanner that finds reachable hosts on a subnet using ICMP ping.

## Usage

```bash
go run *.go 192.168.1.0/24
```

**Note**: Requires sudo for raw ICMP sockets.

## Example Output

```
Scanning subnet: 192.168.1.0/24
Found 254 IPs to scan
Scanning 254 IPs...
Progress: 254/254 (100.0%) - Found 3 hosts

Found 3 reachable hosts:
✓ 192.168.1.1 is reachable
✓ 192.168.1.10 is reachable
✓ 192.168.1.50 is reachable
Scan complete.
```

## Installation

```bash
go build -o neti *.go
```

## Features

- Concurrent scanning with progress indicator
- Consistent sorted output
- Works with any CIDR subnet
