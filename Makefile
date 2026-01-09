.PHONY: build install test coverage lint clean help

# Build variables
BINARY_NAME=pinentry-proton
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/damoun/pinentry-proton/internal/protocol.Version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Install paths
PREFIX?=/usr/local
BINDIR=$(PREFIX)/bin

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GOCLEAN=$(GOCMD) clean

# Default target
all: build

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pinentry-proton

## install: Install the binary to BINDIR (default: /usr/local/bin)
install: build
	install -d $(BINDIR)
	install -m 755 $(BINARY_NAME) $(BINDIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(BINDIR)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Create config: ~/.config/pinentry-proton/config.yaml"
	@echo "  2. Configure GPG: echo 'pinentry-program $(BINDIR)/$(BINARY_NAME)' >> ~/.gnupg/gpg-agent.conf"
	@echo "  3. Reload agent: gpgconf --kill gpg-agent"

## uninstall: Remove the installed binary
uninstall:
	rm -f $(BINDIR)/$(BINARY_NAME)

## test: Run unit tests
test:
	$(GOTEST) -v -race ./...

## test-short: Run unit tests without race detection
test-short:
	$(GOTEST) -v ./...

## test-integration: Run integration tests
test-integration: build
	@./test/run_go_tests.sh

## test-gpg: Run GPG integration tests
test-gpg: build
	@./test/test_gpg.sh

## test-ssh: Run SSH integration tests
test-ssh: build
	@./test/test_ssh.sh

## test-all: Run all tests (unit + integration + GPG + SSH)
test-all: build
	@./test/run_all_tests.sh

## test-setup: Setup test keys (GPG and SSH)
test-setup:
	@./test/setup_test_keys.sh

## coverage: Run tests with coverage report
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

## fmt: Format code
fmt:
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## mod-tidy: Tidy go modules
mod-tidy:
	$(GOMOD) tidy

## mod-download: Download dependencies
mod-download:
	$(GOMOD) download

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.txt coverage.html
	rm -f *.prof

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## release-darwin: Build for macOS
release-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .

## release-linux: Build for Linux
release-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .

## release: Build for all platforms
release: release-darwin release-linux
	@echo "Release builds complete"
	@ls -lh $(BINARY_NAME)-*

## checksums: Generate SHA256 checksums for release builds
checksums:
	@for file in $(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			sha256sum "$$file" > "$$file.sha256"; \
		fi; \
	done
	@echo "Checksums generated"

## pre-commit-install: Install pre-commit hooks (both commit and push)
pre-commit-install:
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit install --hook-type pre-commit; \
		pre-commit install --hook-type pre-push; \
		echo "Pre-commit hooks installed successfully"; \
		echo ""; \
		echo "Hooks configured:"; \
		echo "  - Commit stage: Fast checks (~1-2s)"; \
		echo "  - Push stage: Comprehensive checks (~15-25s)"; \
	else \
		echo "ERROR: pre-commit not found. Install it with:"; \
		echo "  pip install pre-commit"; \
		echo "  or: brew install pre-commit"; \
		exit 1; \
	fi

## pre-commit-run-commit: Run commit-stage hooks manually on all files
pre-commit-run-commit:
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --hook-stage commit --all-files; \
	else \
		echo "ERROR: pre-commit not found. Install it first with: make pre-commit-install"; \
		exit 1; \
	fi

## pre-commit-run-push: Run push-stage hooks manually on all files
pre-commit-run-push:
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --hook-stage push --all-files; \
	else \
		echo "ERROR: pre-commit not found. Install it first with: make pre-commit-install"; \
		exit 1; \
	fi

## pre-commit-run-all: Run all hooks (commit + push stages) manually on all files
pre-commit-run-all:
	@echo "Running commit-stage hooks..."
	@$(MAKE) pre-commit-run-commit
	@echo ""
	@echo "Running push-stage hooks..."
	@$(MAKE) pre-commit-run-push
	@echo ""
	@echo "All pre-commit hooks passed!"
