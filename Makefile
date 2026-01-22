.PHONY: all build clean test install help

# Build variables
GO := go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags="-X main.Version=$(VERSION)"

# Output directories
BIN_DIR := bin
DIST_DIR := dist

all: build

help:
	@echo "torch2pprof - PyTorch to pprof profile converter"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build torch2pprof binary"
	@echo "  install        Install binary to \$$GOPATH/bin"
	@echo "  clean          Remove build artifacts"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  test-race      Run tests with race detector"
	@echo "  fmt            Format code"
	@echo "  vet            Run go vet"
	@echo "  dist           Build for multiple platforms"
	@echo "  help           Show this help message"

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

build: $(BIN_DIR)
	@echo "Building torch2pprof..."
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/torch2pprof ./cmd/torch2pprof
	@echo "Build complete: $(BIN_DIR)/torch2pprof"

install:
	@echo "Installing torch2pprof..."
	$(GO) install ./cmd/torch2pprof
	@echo "Install complete"

clean:
	@echo "Cleaning build artifacts..."
	$(GO) clean
	rm -rf $(BIN_DIR) $(DIST_DIR)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	$(GO) test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race:
	@echo "Running tests with race detector..."
	$(GO) test -v -race ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete"

vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet complete"

# Development targets
dev: fmt vet build

dist: $(BIN_DIR)
	@echo "Building distribution binaries..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/torch2pprof-linux-amd64 ./cmd/torch2pprof
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/torch2pprof-darwin-amd64 ./cmd/torch2pprof
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/torch2pprof-darwin-arm64 ./cmd/torch2pprof
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/torch2pprof-windows-amd64.exe ./cmd/torch2pprof
	@echo "Distribution builds complete: $(DIST_DIR)/"
