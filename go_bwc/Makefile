# Forensic BWC System Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=bwc-system
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe

# Build variables
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test coverage run deps help

all: test build

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v forensic_bwc_system.go

## build-linux: Build for Linux
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v forensic_bwc_system.go

## build-windows: Build for Windows
build-windows:
	@echo "Building for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS) -v forensic_bwc_system.go

## build-mac: Build for macOS
build-mac:
	@echo "Building for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)_mac -v forensic_bwc_system.go

## build-all: Build for all platforms
build-all: build-linux build-windows build-mac
	@echo "Build complete for all platforms"

## test: Run unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 30s ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## test-short: Run quick tests
test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## run: Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run forensic_bwc_system.go

## clean: Clean build files
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_NAME)_mac
	rm -f coverage.out
	rm -f coverage.html
	rm -rf bwc_storage/
	rm -rf test_video.mp4
	rm -rf evidence_export.json
	@echo "Clean complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

## deps-verify: Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

## security: Run security checks (requires gosec)
security:
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "gosec not installed. Run: go install github.com/securego/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec ./...

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t bwc-system:$(VERSION) .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -it --rm -v $(PWD)/bwc_storage:/app/bwc_storage bwc-system:$(VERSION)

## install: Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

## uninstall: Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete"

## help: Display this help message
help:
	@echo "Forensic BWC System - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# Default target
.DEFAULT_GOAL := help
