# Neti Network Scanner Makefile

# Variables
BINARY_NAME=neti
BUILD_DIR=build

# Default target is to show help
.PHONY: all
all: help

# Build for the current platform
.PHONY: build
build:
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all common platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe .
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-macos .
	@echo "All platform builds complete!"
	@ls -la $(BUILD_DIR)/

# Run the scanner with sudo (required for ICMP)
.PHONY: run
run:
	@if [ -z "$(SUBNET)" ]; then \
		echo "Usage: make run SUBNET=192.168.1.0/24"; \
		exit 1; \
	fi
	sudo go run . $(SUBNET)

# Manage dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Show help
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build      Build for the current OS and architecture"
	@echo "  build-all  Build for Linux, Windows, and macOS"
	@echo "  run        Run the scanner (e.g., make run SUBNET=192.168.1.0/24)"
	@echo "  deps       Install and tidy dependencies"
	@echo "  clean      Remove build artifacts"
	@echo "  help       Show this help message"
	@echo ""