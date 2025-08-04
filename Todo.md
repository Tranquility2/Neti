# 📋 Neti - Improvement Roadmap

This document outlines potential improvements and feature additions for the Neti network scanner.

## 🚀 Performance & Scanning Improvements

### 1. Multiple Ping Strategies
- [x] Add TCP SYN scanning as an alternative to ICMP (useful when ICMP is blocked)
- [ ] Implement UDP scanning for specific ports
- [x] Add port scanning capabilities to discovered hosts
- [x] Support for custom port ranges

### 2. Adaptive Timeout & Retry Logic
- [ ] Implement adaptive timeouts based on network latency
- [ ] Add retry mechanism for failed pings with exponential backoff
- [ ] Smart concurrency adjustment based on network conditions
- [ ] Auto-detect optimal concurrency based on system resources

### 3. Enhanced MAC Resolution
- [ ] Implement ARP requests as fallback when cached ARP entries aren't available
- [ ] Add Wake-on-LAN detection for sleeping devices
- [ ] Support for IPv6 neighbor discovery
- [ ] Enhanced ARP table parsing for all platforms

## 📊 Output & User Experience

### 4. Multiple Output Formats
- [ ] JSON output for automation/scripting (`--output json`)
- [ ] CSV export functionality (`--output csv`)
- [ ] XML output support (`--output xml`)
- [ ] Save results to file option (`--save-to filename`)
- [ ] Pretty-print JSON option

### 5. Enhanced Progress Display
- [ ] Show current IP being scanned in progress bar
- [ ] Display scan rate (IPs/second)
- [ ] Add network statistics (avg response time, packet loss)
- [ ] Real-time host discovery counter
- [ ] Improved ETA calculations

### 6. Improved Table Display
- [ ] Color-coded response times (enhance current implementation)
- [ ] Sortable columns (by IP, response time, manufacturer)
- [ ] Filter options (by manufacturer, response time, etc.)
- [ ] Show additional network information (RTT, TTL)
- [ ] Customizable column selection
- [ ] Wide table support for terminal width

## 🔧 Configuration & Flexibility

### 7. Configuration File Support
- [ ] YAML/JSON config file for default settings (`~/.neti/config.yaml`)
- [ ] Profiles for different network types (home, office, datacenter)
- [ ] Custom timeout and concurrency settings per profile
- [ ] Global configuration options

### 8. Advanced Scanning Options
- [ ] Custom port ranges for TCP scanning (`--ports 80,443,22-25`)
- [ ] Exclude IP ranges (`--exclude 192.168.1.1-10`)
- [ ] Include/exclude patterns for hostnames (`--exclude-hostname "*printer*"`)
- [ ] Custom ping packet size and count
- [ ] Scan intensity levels (fast, normal, thorough)

### 9. Network Interface Selection
- [ ] Auto-detect and list available interfaces (`--list-interfaces`)
- [ ] Allow manual interface selection (`--interface eth0`)
- [ ] Support for multiple interfaces simultaneously
- [ ] Interface-specific routing and scanning

## 🛡️ Security & Reliability

### 10. Enhanced Error Handling
- [ ] Graceful handling of permission errors with helpful messages
- [ ] Better network error reporting and categorization
- [ ] Timeout handling improvements with retry suggestions
- [ ] Network connectivity pre-checks

### 11. Privilege Management
- [ ] Check for required privileges before starting scan
- [ ] Fallback options when running without sudo (TCP connect scans)
- [ ] Platform-specific privilege detection and warnings
- [ ] Capability-based permissions on Linux

### 12. Input Validation & Safety
- [ ] Validate subnet ranges to prevent scanning public networks
- [ ] Rate limiting to prevent network flooding
- [ ] Confirmation prompts for large subnet scans
- [ ] Safe defaults for scanning parameters

## 📈 Monitoring & Logging

### 13. Logging System
- [ ] Structured logging with levels (debug, info, warn, error)
- [ ] Log scan history and results to file
- [ ] Debug mode for troubleshooting (`--debug`)
- [ ] Configurable log output (file, stdout, both)

### 14. Statistics & Analytics
- [ ] Scan duration and performance metrics
- [ ] Host availability over time tracking
- [ ] Network health monitoring and trends
- [ ] Historical comparison features

### 15. Scan History
- [ ] Store scan results in local database
- [ ] Compare current scan with previous scans
- [ ] Track new/disappeared hosts
- [ ] Export scan history

## 🔄 Integration & Automation

### 16. API & Automation
- [ ] REST API for remote scanning
- [ ] Webhook notifications for discoveries
- [ ] Integration with network monitoring tools (Nagios, Zabbix)
- [ ] Scheduled scanning with cron-like functionality

### 17. Database Integration
- [ ] SQLite for local storage of scan history
- [ ] PostgreSQL/MySQL support for enterprise use
- [ ] Track changes in network topology over time
- [ ] Database migration and backup tools

### 18. External Tool Integration
- [ ] Integration with nmap for advanced scanning
- [ ] Whois lookup integration
- [ ] GeoIP location detection
- [ ] Vulnerability scanning integration

## 🎯 Code Quality & Architecture

### 19. Code Architecture Improvements
- [ ] Add interfaces for better testability
- [ ] Implement dependency injection pattern
- [ ] Better separation of concerns (MVC pattern)
- [ ] Plugin architecture for extensibility

### 20. Testing & Quality Assurance
- [ ] Unit tests for all components (aim for 90% coverage)
- [ ] Integration tests with mock networks
- [ ] Benchmarking for performance optimization
- [ ] Continuous integration setup

### 21. Performance Optimization
- [ ] Memory usage optimization for large subnets
- [ ] CPU usage profiling and optimization
- [ ] Network bandwidth usage optimization
- [ ] Concurrent processing improvements

## 📚 Documentation & Usability

### 22. Enhanced Documentation
- [ ] Go doc comments for all public functions
- [ ] Usage examples and tutorials
- [ ] API documentation (if REST API is added)
- [ ] Man page creation
- [ ] Video tutorials and demos

### 23. User Experience
- [ ] Interactive mode with menu navigation
- [ ] Guided setup for first-time users
- [ ] Help system with contextual tips
- [ ] Auto-completion for shell integration

### 24. Internationalization
- [ ] Multi-language support for UI messages
- [ ] Localized date/time formatting
- [ ] Cultural considerations for network scanning ethics

## 🔍 Advanced Features

### 25. Network Discovery Enhancement
- [ ] Device type detection (router, printer, phone, etc.)
- [ ] Operating system fingerprinting
- [ ] Service detection on discovered hosts
- [ ] Network topology mapping

### 26. Security Features
- [ ] Stealth scanning modes
- [ ] Honeypot detection
- [ ] Intrusion detection evasion
- [ ] Encrypted communication for API

### 27. Reporting & Visualization
- [ ] Generate PDF/HTML reports
- [ ] Network topology visualization
- [ ] Charts and graphs for scan results
- [ ] Timeline view of network changes

## 🎯 Priority Implementation Recommendations

### Phase 1 (High Priority - Quick Wins)
1. **Multiple output formats** (JSON/CSV export) - Easy to implement, high value
2. **Enhanced error handling** - Improves user experience significantly
3. **Configuration file support** - Enables power user workflows
4. **Input validation & safety** - Prevents misuse and accidents

### Phase 2 (Medium Priority - Core Features)
1. **TCP SYN scanning** - Alternative when ICMP is blocked
2. **Advanced scanning options** - Port ranges, exclusions
3. **Enhanced progress display** - Better user feedback
4. **Logging system** - Essential for debugging and audit

### Phase 3 (Lower Priority - Advanced Features)
1. **Database integration** - For enterprise and historical tracking
2. **REST API** - For automation and integration
3. **Network topology mapping** - Advanced visualization
4. **Plugin architecture** - Long-term extensibility

## 🛠️ Implementation Notes

- Maintain backward compatibility during improvements
- Follow Go best practices and idioms
- Consider cross-platform compatibility for all features
- Prioritize security and ethical scanning practices
- Ensure all new features have appropriate tests
- Document breaking changes and migration paths

---

*Last updated: August 4, 2025*
*This roadmap is a living document and will be updated as features are implemented and new requirements emerge.*
