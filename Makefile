# pdf-cli Makefile

# Variables
BINARY_NAME=pdf
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Directories
BUILD_DIR=build
CMD_DIR=./cmd/pdf

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	GO111MODULE=on $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 GO111MODULE=on $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 GO111MODULE=on $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

.PHONY: build-darwin
build-darwin:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 GO111MODULE=on $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 GO111MODULE=on $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

.PHONY: build-windows
build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 GO111MODULE=on $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

# Install locally
.PHONY: install
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Run tests
.PHONY: test
test:
	GO111MODULE=on $(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	GO111MODULE=on $(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
.PHONY: test-race
test-race:
	GO111MODULE=on $(GOTEST) -race -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Tidy dependencies
.PHONY: tidy
tidy:
	GO111MODULE=on $(GOMOD) tidy

# Download dependencies
.PHONY: deps
deps:
	GO111MODULE=on $(GOMOD) download

# Lint the code (requires golangci-lint)
.PHONY: lint
lint:
	golangci-lint run

# Lint and fix issues
.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix

# Run all checks (comprehensive pre-commit)
.PHONY: check-all
check-all: fmt vet lint test

# Show test coverage percentage
.PHONY: coverage
coverage:
	@GO111MODULE=on $(GOTEST) -coverprofile=coverage.out ./... 2>/dev/null
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

# Check coverage meets threshold (75%)
.PHONY: coverage-check
coverage-check:
	@GO111MODULE=on $(GOTEST) -coverprofile=coverage.out ./... 2>/dev/null
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	if [ $$(echo "$$coverage < 75" | bc -l) -eq 1 ]; then \
		echo "Coverage $$coverage% is below 75% threshold"; \
		exit 1; \
	else \
		echo "Coverage $$coverage% meets 75% threshold"; \
	fi

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Vet code
.PHONY: vet
vet:
	GO111MODULE=on $(GOCMD) vet ./...

# Run all checks
.PHONY: check
check: fmt vet test

# Generate shell completions
.PHONY: completions
completions: build
	mkdir -p completions
	./$(BINARY_NAME) completion bash > completions/$(BINARY_NAME).bash
	./$(BINARY_NAME) completion zsh > completions/_$(BINARY_NAME)
	./$(BINARY_NAME) completion fish > completions/$(BINARY_NAME).fish

# Show version
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Help
.PHONY: help
help:
	@echo "pdf-cli Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build the binary"
	@echo "  make build        Build the binary"
	@echo "  make build-all    Build for all platforms"
	@echo "  make install      Install to GOPATH/bin"
	@echo "  make test         Run tests"
	@echo "  make test-coverage Run tests with coverage"
	@echo "  make clean        Clean build artifacts"
	@echo "  make tidy         Tidy dependencies"
	@echo "  make deps         Download dependencies"
	@echo "  make lint         Run linter"
	@echo "  make lint-fix     Run linter with auto-fix"
	@echo "  make fmt          Format code"
	@echo "  make vet          Vet code"
	@echo "  make check        Run all checks (fmt, vet, test)"
	@echo "  make check-all    Run all checks (fmt, vet, lint, test)"
	@echo "  make coverage     Show test coverage percentage"
	@echo "  make coverage-check Check coverage meets 75% threshold"
	@echo "  make completions  Generate shell completions"
	@echo "  make version      Show version info"
	@echo "  make help         Show this help"
