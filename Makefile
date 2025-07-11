# Neti Network Scanner Makefile

# Variables
BINARY_NAME=neti
GO_FILES=$(wildcard *.go)
BUILD_DIR=build

# Platform-specific binary names
BINARY_LINUX=$(BINARY_NAME)
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_MACOS=$(BINARY_NAME)-macos

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for Linux (default)
.PHONY: build-linux
build-linux: $(BUILD_DIR)/$(BINARY_LINUX)

$(BUILD_DIR)/$(BINARY_LINUX): $(GO_FILES)
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_LINUX) .
	@echo "Linux build complete: $(BUILD_DIR)/$(BINARY_LINUX)"

# Build for Windows
.PHONY: build-windows
build-windows: $(BUILD_DIR)/$(BINARY_WINDOWS)

$(BUILD_DIR)/$(BINARY_WINDOWS): $(GO_FILES)
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_WINDOWS) .
	@echo "Windows build complete: $(BUILD_DIR)/$(BINARY_WINDOWS)"

# Build for macOS
.PHONY: build-macos
build-macos: $(BUILD_DIR)/$(BINARY_MACOS)

$(BUILD_DIR)/$(BINARY_MACOS): $(GO_FILES)
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_MACOS) .
	@echo "macOS build complete: $(BUILD_DIR)/$(BINARY_MACOS)"

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-windows build-macos
	@echo "All platform builds complete!"
	@ls -la $(BUILD_DIR)/

# Run the program (requires subnet argument)
.PHONY: run
run:
	@if [ -z "$(SUBNET)" ]; then \
		echo "Usage: make run SUBNET=192.168.1.0/24"; \
		exit 1; \
	fi
	go run . $(SUBNET)

# Run with sudo (for ICMP sockets)
.PHONY: run-sudo
run-sudo:
	@if [ -z "$(SUBNET)" ]; then \
		echo "Usage: make run-sudo SUBNET=192.168.1.0/24"; \
		exit 1; \
	fi
	sudo go run . $(SUBNET)

# Install dependencies
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

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Run tests
.PHONY: test
test:
	go test ./...

# Install binary to system
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Uninstall binary from system
.PHONY: uninstall
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "Development setup complete"

# Show help
.PHONY: help
help:
	@echo "Neti Network Scanner - Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build the binary for current platform"
	@echo "  build-linux    - Build for Linux (amd64)"
	@echo "  build-windows  - Build for Windows (amd64)"
	@echo "  build-macos    - Build for macOS (amd64)"
	@echo "  build-all      - Build for all platforms"
	@echo ""
	@echo "Run targets:"
	@echo "  run            - Run the program (use: make run SUBNET=192.168.1.0/24)"
	@echo "  run-sudo       - Run with sudo (use: make run-sudo SUBNET=192.168.1.0/24)"
	@echo ""
	@echo "Other targets:"
	@echo "  deps           - Install dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  test           - Run tests"
	@echo "  install        - Install binary to /usr/local/bin"
	@echo "  uninstall      - Remove binary from /usr/local/bin"
	@echo "  dev-setup      - Setup development environment"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build-all                    # Build for all platforms"
	@echo "  make run-sudo SUBNET=192.168.1.0/24"
	@echo "  make build-windows                # Build Windows executable"