# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pinentry-Proton is a secure pinentry program that integrates ProtonPass with GPG and SSH agents. It implements the Assuan pinentry protocol to retrieve passwords from ProtonPass vaults using the official ProtonPass CLI (`pass-cli`), eliminating the need to manually type passphrases while maintaining security.

**Core Purpose**: Act as a pinentry replacement that retrieves passwords from ProtonPass instead of prompting the user interactively.

**Primary Use Case**: Enable signing git commits (and other GPG/SSH operations) without repeated PIN entry. User unlocks ProtonPass UI once at session start, then pinentry-proton automatically fetches PINs for GPG operations like commit signing - eliminating repetitive manual PIN entry.

## Essential Commands

### Build & Development
```bash
# Build the binary
make build

# Install to /usr/local/bin (requires sudo)
sudo make install

# Run all quality checks (format, vet, lint, test)
make check

# Format code
make fmt

# Run linters (requires golangci-lint)
make lint
```

### Testing
```bash
# Run unit tests (internal packages only)
make test-unit

# Run unit tests with race detection
go test -race ./internal/...

# Run E2E tests (mock ProtonPass, no auth needed)
make test-e2e

# Run tests with real ProtonPass (requires auth)
make test-realpass

# Run integration tests (requires built binary)
make test-integration

# Run GPG integration tests (requires GPG, pass-cli, test keys)
make test-gpg

# Run SSH integration tests (requires SSH tools, pass-cli, test keys)
make test-ssh

# Run ALL tests (unit + integration + GPG + SSH)
make test-all

# Setup test keys for GPG and SSH
make test-setup

# Generate coverage report with threshold check (75%)
make test-coverage-check

# Run CI test suite (unit + coverage check)
make test-ci

# Run benchmarks
make benchmark

# Save benchmark baseline
make benchmark-save

# Compare with baseline (requires benchstat)
make benchmark-compare
```

### Running Single Tests
```bash
# Test a specific package
go test -v ./internal/config
go test -v ./internal/protocol

# Test a specific function
go test -v -run TestSpecificFunction ./internal/config
```

### Pre-Commit Hooks
```bash
# Install pre-commit hooks (first time setup)
make pre-commit-install

# Run commit-stage hooks manually (fast checks)
make pre-commit-run-commit

# Run push-stage hooks manually (comprehensive checks)
make pre-commit-run-push

# Run all hooks manually
make pre-commit-run-all

# Bypass hooks when needed (use sparingly)
git commit --no-verify
git push --no-verify
```

## Pre-Commit Hook System

This project uses [pre-commit](https://pre-commit.com/) with hooks split between commit and push stages for optimal developer experience.

### Hook Stages

**Commit Stage (~1-2s)** - Fast feedback on every `git commit`:
- `go fmt` - Code formatting
- `go vet` - Static analysis
- `go test` - Unit tests (WITHOUT race detection)
- `go mod verify` - Module integrity
- Branch protection - Prevent commits to main

**Push Stage (~15-25s)** - Comprehensive validation on every `git push`:
- All commit-stage checks (redundant safety net)
- `go test -race` - Unit tests WITH race detection
- `golangci-lint` - Full linting (11 linters, no --fast flag)
- `go mod tidy` - Verify modules are tidy
- `make build` - Build verification
- Secrets scan - Check for accidentally committed secrets

### Design Philosophy

**Why split between commit and push?**
- Commit hooks provide fast feedback (<2s) without interrupting flow
- Race detection adds 4.5s to test execution, moved to push stage
- Push hooks mirror CI/CD pipeline for consistency
- Developers get quick iteration cycles while maintaining quality gates

### Installation

First-time setup for developers:
```bash
# Install pre-commit tool (if not already installed)
pip install pre-commit
# or
brew install pre-commit

# Install hooks for this repository
make pre-commit-install
```

This installs:
- `.git/hooks/pre-commit` - Runs commit-stage hooks
- `.git/hooks/pre-push` - Runs push-stage hooks

### Manual Hook Execution

Run hooks without committing/pushing:
```bash
# Test what would run on commit
make pre-commit-run-commit

# Test what would run on push
make pre-commit-run-push

# Run everything
make pre-commit-run-all
```

Or use pre-commit directly:
```bash
pre-commit run --hook-stage commit --all-files
pre-commit run --hook-stage push --all-files
```

### Bypassing Hooks

In rare cases when you need to skip hooks:
```bash
git commit --no-verify    # Skip commit hooks
git push --no-verify      # Skip push hooks
```

**When to bypass:**
- Emergency hotfixes
- WIP commits on personal branches
- When hooks are broken/misconfigured
- When committing pre-commit infrastructure itself

**Important**: Use sparingly. Hooks exist to catch issues before they reach CI/CD.

### GitHub Actions Integration

The `.github/workflows/ci.yml` includes a `pre-commit` job that:
- Runs all push-stage hooks on every PR and push to main
- Provides safety net for bypassed local hooks
- Ensures consistency between local development and CI
- Fails CI if any hook fails

This means:
- Contributors without local hooks installed are still validated
- Bypassed hooks (`--no-verify`) are caught in CI
- Same quality checks run locally and remotely

### Configuration

Hook configuration is in `.pre-commit-config.yaml`:
- Uses `local` repo to run project-specific commands
- Hooks use `language: system` (no isolated environments)
- Commit hooks have `stages: [pre-commit]`
- Push hooks have `stages: [pre-push]`

### Optional: golangci-lint on Commit

By default, full `golangci-lint` only runs on push. To enable faster linting on commit:

1. Uncomment the `golangci-lint-fast` hook in `.pre-commit-config.yaml`
2. Reinstall hooks: `make pre-commit-install`

This adds 3-5s to commit time but catches more issues early.

### Common Issues and Solutions

**Issue**: Hooks not running
```bash
# Solution: Ensure hooks are installed
make pre-commit-install
```

**Issue**: Tests failing in hooks but passing manually
```bash
# Solution: Hooks run from repo root, check working directory
# Debug: Add echo $PWD to hook entry in .pre-commit-config.yaml
```

**Issue**: golangci-lint not found
```bash
# Solution: Install golangci-lint
brew install golangci-lint
# or follow: https://golangci-lint.run/usage/install/

# Or let the hook skip gracefully (already configured)
```

**Issue**: Hooks too slow
```bash
# Solution 1: Check if golangci-lint --fast is enabled on commit
# Solution 2: Run only changed files (but current config runs all files for consistency)
# Solution 3: Temporarily bypass with --no-verify (not recommended)
```

**Issue**: Pre-commit tool not found
```bash
# Solution: Install pre-commit
pip install pre-commit    # via pip
brew install pre-commit   # via Homebrew
```

### Makefile Targets

The Makefile includes convenience targets for hook management:

- `pre-commit-install`: Install both pre-commit and pre-push hooks
- `pre-commit-run-commit`: Run commit-stage hooks manually
- `pre-commit-run-push`: Run push-stage hooks manually
- `pre-commit-run-all`: Run all hooks (commit + push) manually

These targets check if pre-commit is installed and provide helpful error messages if not.

### CI/CD Parity

The push-stage hooks match the CI pipeline:
- ✅ `go fmt` → CI validates formatting (via go vet)
- ✅ `go vet` → CI test job
- ✅ `go test -race` → CI test job with coverage
- ✅ `golangci-lint` (full) → CI lint job
- ✅ `make build` → CI build job
- ✅ `go mod verify` → CI test job
- ✅ Secrets scan → Local only (additional protection)

This ensures that if push hooks pass, CI should pass (barring environment differences).

### Performance

Measured execution times on this project:
- `go fmt ./...`: ~0.2s
- `go vet ./...`: ~0.5s
- `go test ./...` (no race): ~0.4s
- `go mod verify`: ~0.02s
- Branch protection check: ~0.01s
- **Total commit stage**: ~1.1s ✅

- `go test -race ./...`: ~4.5s
- `golangci-lint run ./...`: ~5-15s
- `go mod tidy` check: ~0.5s
- `make build`: ~0.2s
- Secrets scan: ~0.1s
- **Total push stage**: ~11-21s ✅

### Best Practices

1. **Always install hooks**: First thing after cloning the repo
2. **Don't bypass regularly**: If you're bypassing often, hooks may be too strict
3. **Fix issues immediately**: Don't let hook failures accumulate
4. **Test manually**: Run `make pre-commit-run-all` before pushing for confidence
5. **Keep hooks fast**: Commit stage should stay under 2s for good DX
6. **Update documentation**: If you modify hooks, update this section

## Architecture Overview

### Package Structure

```
cmd/pinentry-proton/          Entry point, signal handling, orchestration
internal/config/              Configuration loading and context matching
internal/protocol/            Pinentry Assuan protocol implementation
internal/protonpass/          ProtonPass CLI integration
internal/platform/            Platform-specific code (macOS/Linux)
```

### Key Architectural Concepts

**1. Protocol Flow**
- GPG/SSH agent sends pinentry commands via stdin
- `protocol.Session` parses Assuan protocol commands
- Context fields (description, title, keyinfo) are collected via SET* commands
- On GETPIN, config matcher finds appropriate ProtonPass item
- ProtonPass client retrieves password via `pass-cli item get`
- Password is percent-encoded and returned via protocol
- All sensitive data is zeroed from memory

**2. Configuration Matching System**
The matcher in `internal/config/config.go` uses case-insensitive substring matching:
- `FindItemForContext()` iterates through mappings in order
- Each mapping checks if ALL specified criteria match (AND logic)
- First match wins; falls back to `default_item` if no match
- Match criteria: description, prompt, title, keyinfo

**3. Security Architecture**
- **Memory Safety**: All passwords tracked in `Session.sensitiveData[]` and zeroed via `protonpass.ZeroBytes()`
- **No Logging**: Passwords never appear in logs (debug mode only logs metadata)
- **No Persistence**: Passwords never written to disk
- **Signal Handling**: SIGINT/SIGTERM trigger cleanup before exit
- **Context Cancellation**: All operations respect context timeouts

**4. ProtonPass Integration**
- Uses `pass-cli item get <vault/item> --field <field>` command
- Parses URI format: `pass://VAULT/ITEM/FIELD`
- Default field is `password` if not specified
- No passwords in command-line args (retrieved from stdout)

## Critical Security Requirements

When modifying this codebase, you MUST:

1. **Never log passwords**: Even in debug mode, only log metadata (item URIs, lengths, context info)
2. **Zero sensitive data**: Call `protonpass.ZeroBytes()` on all password byte slices when done
3. **Track for cleanup**: Add password buffers to `Session.sensitiveData` for automatic cleanup
4. **No string conversion**: Keep passwords as `[]byte` to enable zeroing (strings are immutable)
5. **Validate input**: Check all configuration, URIs, and protocol input for malformed data
6. **Handle signals**: Ensure cleanup happens on SIGINT/SIGTERM
7. **Respect context**: All long-running operations should accept and check `context.Context`

See SECURITY.md for comprehensive security policy and threat model.

## Testing Architecture

### Test Suite Overview

The project has **comprehensive test coverage (82.4%)**  with multiple test layers:

- **Unit Tests** - Package-level tests with mocks (protonpass: 88.2%, platform: 100%)
- **E2E Tests** - Mock-based application tests (no ProtonPass auth required, runs in CI)
- **Integration Tests** - Binary protocol tests with real pinentry commands
- **Benchmarks** - Performance baselines and regression detection
- **GPG/SSH Tests** - Real cryptographic operations (optional, requires ProtonPass auth)

### Coverage Enforcement

**Minimum coverage: 75%** - Enforced in CI and `make test-coverage-check`

Current coverage:
- Overall: **82.4%**
- ProtonPass: 88.2%
- Platform: 100%
- Config: 87.1%
- Protocol: 77.4%

### Test Requirements

- **Test PIN**: All tests use `424242` as the test passphrase
- **Mock Infrastructure**: E2E tests use sophisticated mock pass-cli (no ProtonPass needed)
- **Optional Real Tests**: Use `-tags=realpass` for tests with real ProtonPass
- **Test Vault**: Real ProtonPass tests require `pass://test/pinentry-code/password`

### Test Structure

```
test/
├── testutil/                # Shared test utilities
│   ├── fixtures.go          # Config helpers, assertions, setup
│   └── mock_pass.go         # Mock ProtonPass CLI implementation
├── e2e/                     # End-to-end tests (mock-based, CI-friendly)
│   ├── e2e_test.go          # Complete workflows without ProtonPass
│   └── e2e_realpass_test.go # Optional tests with real pass-cli
├── fixtures/                # Test data, keys, mock configs
├── integration_test.go      # Binary protocol tests
├── test_gpg.sh             # GPG workflow (requires ProtonPass)
├── test_ssh.sh             # SSH workflow (requires ProtonPass)
└── run_all_tests.sh        # Master test runner

internal/*/
├── *_test.go               # Unit tests (table-driven, comprehensive)
└── *_benchmark_test.go     # Performance benchmarks
```

### Test Keys

- Located in `test/fixtures/gnupg/` (GPG) and `test/fixtures/ssh/` (SSH)
- All use passphrase `424242`
- **Never use in production** - committed for testing only

### Writing Tests

**Unit Tests:**
- Use table-driven tests with subtests (`t.Run()`)
- Mock external dependencies (ProtonPass CLI)
- Test error cases and edge conditions
- Ensure 75%+ coverage for new code

**E2E Tests:**
- Use shared `test/testutil` utilities
- Create isolated test environments (temp dirs, mock CLIs)
- Test complete protocol flows
- Verify cleanup and memory zeroing

**Benchmarks:**
- Use `testing.B` interface
- Include memory allocation stats (`-benchmem`)
- Test at multiple scales (small, medium, large data)
- Save baselines with `make benchmark-save`

### CI/CD Integration

Tests run automatically on GitHub Actions:
- Matrix: Ubuntu + macOS, Go 1.21 + 1.22
- Unit tests with race detection
- E2E tests (no ProtonPass needed)
- Coverage enforcement (fails if <75%)
- Benchmark tracking (artifacts stored for 30 days)

See `TESTING.md` for detailed testing guide.

## Common Development Patterns

### Adding a New Protocol Command

1. Add case to `handleCommand()` in `internal/protocol/session.go`
2. Implement handler function (e.g., `handleNewCommand()`)
3. Write unit test in `internal/protocol/protocol_test.go`
4. Add integration test in `test/integration_test.go`

### Adding a New Configuration Option

1. Add field to `Config` struct in `internal/config/config.go`
2. Update `Validate()` method
3. Add to `config.example.yaml`
4. Update README.md with documentation
5. Add unit tests in `internal/config/config_test.go`

### Modifying Password Retrieval

When touching `internal/protonpass/client.go`:
- Passwords must remain as `[]byte`
- Call `ZeroBytes()` in defer immediately after getting password
- Never log the password content
- Trim whitespace from pass-cli output
- Return descriptive errors without including password

## Platform Considerations

### macOS
- SSH agent integration may prefer system Keychain
- GPG agent works normally with pinentry-proton
- Optional Keychain integration in `internal/platform/platform_darwin.go`

### Linux
- Full SSH and GPG agent support
- Use `termios` for secure terminal input
- Optional libsecret integration in `internal/platform/platform_linux.go`

### Build Tags
Platform-specific code uses build tags:
```go
//go:build darwin
//go:build linux
//go:build !darwin && !linux
```

## Configuration System

### Config File Locations (checked in order)
1. `$PINENTRY_PROTON_CONFIG`
2. `$XDG_CONFIG_HOME/pinentry-proton/config.yaml`
3. `$HOME/.config/pinentry-proton/config.yaml`
4. `$HOME/.pinentry-proton.yaml`

### ProtonPass URI Format
```
pass://VAULT_NAME/ITEM_TITLE/FIELD
pass://SHARE_ID/ITEM_ID/FIELD
```
- Field is optional (defaults to `password`)
- Examples: `pass://Work/SSH Key/password`, `pass://Personal/GPG/passphrase`

## Debug Mode

Enable verbose logging:
```bash
export PINENTRY_PROTON_DEBUG=1
./pinentry-proton
```

This logs:
- Protocol commands received
- Context values (desc, prompt, title, keyinfo)
- Matched ProtonPass item URI
- Password length (NOT password content)
- pass-cli command execution and errors

## Important Files

- `ARCHITECTURE.md`: Detailed architecture, design principles, data flow
- `SECURITY.md`: Security policy and threat model
- `TESTING.md`: Comprehensive testing guide
- `CONTRIBUTING.md`: Contribution guidelines and code review checklist
- `config.example.yaml`: Example configuration with all options

## Dependencies

Minimal dependency policy - prefer standard library:
- `gopkg.in/yaml.v3`: YAML config parsing
- Standard library for everything else

When adding dependencies:
1. Justify necessity in PR
2. Check for CVEs and maintenance status
3. Run `go mod verify` and `go mod tidy`
4. Update this documentation

## Common Pitfalls

1. **Don't convert passwords to strings**: Use `[]byte` to enable memory zeroing
2. **Don't forget cleanup**: Always defer `protonpass.ZeroBytes()` and `session.Cleanup()`
3. **Don't log sensitive data**: Check all log statements, even in debug mode
4. **Don't use relative imports**: All imports should use full module path
5. **Don't skip race detection**: Run `go test -race` before submitting
6. **Don't commit secrets**: Even in tests, use mock credentials only

## Release Process

Version is set in `internal/protocol/session.go`:
```go
const Version = "1.0.0"
```

Build commands inject version from git:
```bash
# Makefile automatically sets version from git describe
make build
```
