# Neti Network Scanner Makefile

# Variables
BINARY_NAME=neti
GO_FILES=$(wildcard *.go)
BUILD_DIR=build

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_FILES)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the program (requires subnet argument)
.PHONY: run
run:
	@if [ -z "$(SUBNET)" ]; then \
		echo "Usage: make run SUBNET=192.168.1.0/24"; \
		exit 1; \
	fi
	go run $(GO_FILES) $(SUBNET)

# Run with sudo (for ICMP sockets)
.PHONY: run-sudo
run-sudo:
	@if [ -z "$(SUBNET)" ]; then \
		echo "Usage: make run-sudo SUBNET=192.168.1.0/24"; \
		exit 1; \
	fi
	sudo go run $(GO_FILES) $(SUBNET)

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
	@echo "  build      - Build the binary"
	@echo "  run        - Run the program (use: make run SUBNET=192.168.1.0/24)"
	@echo "  run-sudo   - Run with sudo (use: make run-sudo SUBNET=192.168.1.0/24)"
	@echo "  deps       - Install dependencies"
	@echo "  clean      - Clean build artifacts"
	@echo "  fmt        - Format code"
	@echo "  test       - Run tests"
	@echo "  install    - Install binary to /usr/local/bin"
	@echo "  uninstall  - Remove binary from /usr/local/bin"
	@echo "  dev-setup  - Setup development environment"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make run SUBNET=192.168.1.0/24"
	@echo "  make run-sudo SUBNET=10.0.0.0/16"