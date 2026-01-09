# Testing Guide

This document describes the comprehensive test suite for pinentry-proton.

## Overview

A comprehensive test suite covering unit tests, protocol integration, and real-world GPG/SSH workflows. All tests use PIN **424242** and the ProtonPass test item URI (see Test Configuration below).

### Test Suite Summary

**Test Types:**
- **Unit Tests** - Go package tests for config, protocol, protonpass, and platform packages
- **End-to-End Tests** - Complete application flow tests with mock ProtonPass (no authentication required)
- **Integration Tests** - Binary protocol tests using real pinentry commands
- **Benchmark Tests** - Performance baseline and regression testing
- **GPG Integration** - Real GPG signing, verification, encryption, and decryption
- **SSH Integration** - Real SSH key operations and agent interaction

**Test Infrastructure:**
- **566 lines** of unit tests for protonpass package (88.2% coverage)
- **120 lines** of platform tests (100% coverage)
- **639 lines** of E2E tests with sophisticated mock infrastructure
- **526 lines** of benchmark tests across all packages
- **Overall coverage: 82.4%** (target: 75%)
- 8 Go integration tests covering complete pinentry protocol
- GPG/SSH end-to-end tests with real cryptographic operations
- Shared test utilities in `test/testutil/` for reusable mock infrastructure
- Mock ProtonPass CLI implementation for CI/CD testing
- Test key infrastructure (GPG + SSH with passphrase 424242)
- Comprehensive Makefile targets for easy test execution
- CI/CD pipeline with coverage enforcement

**Quick Start:**
```bash
make build          # Build the binary
make test-setup     # Create test keys (first time only)
make test-all       # Run all tests
```

## Test Configuration

### Test PIN

All tests use PIN: **424242**

### Test ProtonPass Item

Default test item URI:
```
pass://KmDQwh8YtmA3hKFn1x4KucB4ZXBG4_GXKLKp9oRP6uf_jn8wTTjzjnnP7A92KdQXmLp4kvgBAertdUZgggtZhQ==/MYhqRQ1mT5yo-l0TUh_Dzm38QvCsegOdKU2OWemXRheOOVAuv46qq7UBf6gWX3ZfiMDoOKnlfpSSPzAKRR_BRg==
```

This item must exist in your ProtonPass vault and contain the password `424242`.

## Quick Start

### Run All Tests

```bash
make test-all
```

This runs:
- Unit tests
- Integration tests  
- GPG tests
- SSH tests

### Run Individual Test Suites

```bash
# Unit tests only
make test

# Integration tests (binary protocol)
make test-integration

# GPG integration
make test-gpg

# SSH integration  
make test-ssh
```

## Test Setup

### Initial Setup

Create test keys (GPG and SSH):

```bash
make test-setup
```

This creates:
- `test/fixtures/gnupg/` - GPG home with test key
- `test/fixtures/ssh/` - SSH test keys
- Both use passphrase: `424242`

### Manual Setup

If you want to create keys manually:

```bash
cd test
./setup_test_keys.sh
```

## Test Structure

```
test/
├── fixtures/                 # Test data and keys
│   ├── gnupg/               # GPG test keyring
│   ├── ssh/                 # SSH test keys
│   ├── test-config.yaml     # Test configuration
│   └── README.md            # Fixture documentation
├── testutil/                # Shared test utilities
│   ├── fixtures.go          # Test helpers (config, assertions, setup)
│   └── mock_pass.go         # Sophisticated mock ProtonPass CLI
├── e2e/                     # End-to-end tests (no ProtonPass required)
│   ├── e2e_test.go          # Mock-based E2E tests (639 lines)
│   └── e2e_realpass_test.go # Optional real ProtonPass tests (build tag: realpass)
├── integration_test.go      # Go integration tests
├── setup_test_keys.sh       # Key creation script
├── test_gpg.sh             # GPG integration test
├── test_ssh.sh             # SSH integration test
├── run_go_tests.sh         # Go test runner
└── run_all_tests.sh        # Master test runner

internal/
├── config/
│   ├── config_test.go           # Config unit tests
│   └── config_benchmark_test.go # Config benchmarks (252 lines)
├── protocol/
│   ├── protocol_test.go              # Protocol unit tests
│   ├── integration_test.go           # Protocol integration tests
│   └── protocol_benchmark_test.go    # Protocol benchmarks (144 lines)
├── protonpass/
│   ├── client_test.go               # ProtonPass unit tests (566 lines)
│   └── client_benchmark_test.go     # ProtonPass benchmarks (130 lines)
└── platform/
    └── platform_test.go             # Platform tests (120 lines)
```

## Unit Tests

Location: `internal/*/`

Tests the core packages:

```bash
go test ./internal/config
go test ./internal/protocol
go test ./internal/protonpass
```

Coverage:
- Configuration loading and validation
- Pattern matching for item selection
- Protocol command parsing
- Encoding/decoding utilities
- Memory zeroing
- Signal handling
- ProtonPass client (88.2% coverage)
- Platform abstraction (100% coverage)

## End-to-End Tests

Location: `test/e2e/`

**True end-to-end tests that work in CI/CD without ProtonPass authentication.**

These tests use sophisticated mock infrastructure to validate complete application workflows.

### Key Features

- **No ProtonPass Required** - Uses mock pass-cli implementation
- **Fast Execution** - All E2E tests run in ~3 seconds
- **Comprehensive Coverage** - Tests complete protocol flows, error handling, edge cases
- **CI/CD Friendly** - Runs on GitHub Actions without authentication

### Test Scenarios

1. **TestE2E_FullProtocolFlow** - Complete protocol validation (SETDESC → GETPIN → BYE)
2. **TestE2E_MultipleRequestsSameSession** - Multiple GETPIN requests in one session
3. **TestE2E_LongPassword** - Passwords >1KB (tested with 2KB)
4. **TestE2E_SpecialCharacters** - Unicode, symbols, quotes, spaces (newlines excluded - protocol limitation)
5. **TestE2E_GPGWorkflow** - GPG signing context matching
6. **TestE2E_SSHWorkflow** - SSH key unlock context matching

### Running E2E Tests

```bash
# Build first (E2E tests need the binary)
make build

# Run E2E tests
make test-e2e
# or
go test -v ./test/e2e/
```

### Optional: Real ProtonPass Tests

For developers with authenticated ProtonPass:

```bash
# Run tests with real pass-cli (requires auth)
make test-realpass
# or
go test -v -tags=realpass ./test/e2e/
```

**Requirements:**
- Authenticated pass-cli
- Test vault item: `pass://test/pinentry-code/password`
- Password must be set in vault

### Mock Infrastructure

The E2E tests use shared utilities from `test/testutil/`:

- **MockPassCLI** - Sophisticated mock with call tracking, latency simulation, failure injection
- **Test Helpers** - Config generation, protocol assertions, environment setup
- **Wrapper Scripts** - Ensures mock CLI is used instead of system pass-cli

## Benchmark Tests

Location: `internal/*/client_benchmark_test.go`, `internal/*/protocol_benchmark_test.go`, `internal/*/config_benchmark_test.go`

Performance testing and regression detection.

### Running Benchmarks

```bash
# Run all benchmarks
make benchmark

# Save baseline
make benchmark-save

# Compare with baseline (requires benchstat)
make benchmark-compare
```

### Benchmark Coverage

**ProtonPass Package:**
- Password retrieval (~3ms with mock)
- Long password handling (1KB)
- Special character handling
- URI parsing
- Memory zeroing (0.95ns for 16B to 3.7µs for 1MB)

**Protocol Package:**
- Percent encoding at various sizes (8B to 16KB)
- Encoding throughput: ~500 MB/s for large data
- Escape/unescape operations
- Round-trip encode/decode
- Session reset operations

**Config Package:**
- Config loading (23µs small, 290µs for 100 mappings)
- Context matching (O(1) for first match, O(n) for scanning)
- Pattern matching (case-insensitive, wildcards)
- Validation logic

### Performance Insights

```
BenchmarkZeroBytes/16B-14      1000000000   0.95 ns/op   16826 MB/s
BenchmarkZeroBytes/1MB-14         638046   3759 ns/op   278925 MB/s
BenchmarkPercentEncode/1KB-14    1000000   2009 ns/op      509 MB/s
BenchmarkFindItemForContext-14  26414796     92 ns/op       32 B/op
```

## Integration Tests

Location: `test/integration_test.go`

Tests the complete binary using the Assuan protocol:

### Test Cases

1. **TestBinaryExists** - Verify build
2. **TestBasicProtocol** - Handshake and GETINFO commands
3. **TestSetAndGetOptions** - SETDESC, SETPROMPT, etc.
4. **TestGetPinWithMockProtonPass** - GETPIN with mock pass CLI
5. **TestGetPinWithGPGContext** - GPG-specific context
6. **TestCancelGetPin** - Error handling
7. **TestInvalidCommands** - Invalid command handling
8. **TestMultipleGetPin** - Multiple requests in one session

### Running

```bash
make test-integration
# or
go test -v ./test
```

## GPG Integration Tests

Location: `test/test_gpg.sh`

End-to-end test with real GPG operations.

### What It Tests

1. **Signing** - Sign a message using GPG with pinentry-proton
2. **Verification** - Verify the signature
3. **Encryption** - Encrypt a message
4. **Decryption** - Decrypt and verify content

### Requirements

- GPG installed (`gpg` command)
- Test keys created (`make test-setup`)
- ProtonPass CLI authenticated
- Test item exists with PIN 424242

### Running

```bash
make test-gpg
# or
./test/test_gpg.sh
```

### Process Flow

1. Configure GPG to use pinentry-proton
2. Kill existing gpg-agent
3. Sign message (triggers pinentry)
4. Verify signature
5. Encrypt message
6. Decrypt message (triggers pinentry)
7. Verify content matches

### Expected Output

```
==> GPG Integration Test

Using GPG key: <KEY_ID>

Test 1: Sign a message
✓ Message signed successfully

Test 2: Verify signature
✓ Signature verified successfully

Test 3: Encrypt and decrypt
✓ Message encrypted
✓ Message decrypted
✓ Decrypted content matches original

==> All GPG tests passed! 🎉
```

## SSH Integration Tests

Location: `test/test_ssh.sh`

End-to-end test with SSH operations.

### What It Tests

1. **SSH Agent** - Start ssh-agent
2. **Add Key** - Add password-protected key using pinentry
3. **List Keys** - Verify key loaded
4. **Sign Data** - Use key to sign data (if supported)

### Requirements

- SSH tools (`ssh-agent`, `ssh-add`, `ssh-keygen`)
- Test keys created (`make test-setup`)
- ProtonPass CLI authenticated
- Test item exists with PIN 424242

### Running

```bash
make test-ssh
# or
./test/test_ssh.sh
```

### Platform Notes

**macOS:** SSH may prefer system keychain over SSH_ASKPASS. The test verifies pinentry-proton can retrieve passphrases even if full SSH agent integration has platform limitations.

**Linux:** Should work with SSH_ASKPASS environment variable.

### Expected Output

```
==> SSH Integration Test

Using SSH key: test/fixtures/ssh/test_ed25519

Test 1: Start SSH agent
✓ SSH agent started

Test 2: Add SSH key using pinentry-proton
✓ SSH key added to agent

Test 3: List loaded keys
✓ Test key is loaded in agent

Test 4: Test key usage (sign data)
✓ Successfully signed data with SSH key
✓ Signature verified successfully

==> SSH tests completed! 🎉
```

## Mock ProtonPass CLI

For testing without ProtonPass, the integration tests create a mock `pass` binary:

```bash
#!/bin/sh
echo "424242"
exit 0
```

This allows testing the pinentry protocol without requiring ProtonPass authentication.

## Continuous Integration

The test suite is fully integrated with GitHub Actions CI/CD pipeline.

### CI Pipeline Features

- **Matrix Testing** - Tests on Ubuntu and macOS with Go 1.21 and 1.22
- **Unit Tests** - All internal packages with race detection
- **E2E Tests** - Mock-based tests (no ProtonPass authentication needed)
- **Coverage Enforcement** - Fails if coverage drops below 75%
- **Benchmarks** - Automated performance baseline tracking
- **Pre-commit Hooks** - Format, vet, and lint checks
- **Security Scanning** - Gosec vulnerability detection

### GitHub Actions Workflow

The CI workflow (`.github/workflows/ci.yml`) runs:

1. **Pre-commit Hooks Job** - Fast checks (format, vet, lint)
2. **Test Job** - Unit + E2E tests with coverage
3. **Lint Job** - golangci-lint comprehensive checks
4. **Build Job** - Multi-platform binary builds
5. **Security Job** - Gosec security scanner
6. **Benchmark Job** - Performance regression tracking

### Coverage Enforcement

Coverage is automatically checked in CI:

```bash
# CI runs this check
make test-coverage-check

# Fails if coverage < 75%
ERROR: Coverage 72.5% is below minimum 75%
```

Current coverage: **82.4%** (✅ above 75% threshold)

### Running CI Tests Locally

```bash
# Run the same tests as CI
make test-ci

# This runs:
# - Unit tests with race detection
# - Coverage generation
# - Coverage threshold check
```

### Benchmark Tracking

Benchmarks run automatically on every CI build:

- Results stored as artifacts (30-day retention)
- Compare performance across commits
- Detect performance regressions

**Important:** E2E tests use mock infrastructure, so **no ProtonPass authentication is needed** in CI.

## Test Coverage

Generate coverage report:

```bash
make coverage
```

Opens `coverage.html` showing line-by-line coverage.

## Debugging Tests

### Enable Debug Mode

```bash
export PINENTRY_PROTON_DEBUG=1
make test-gpg
```

This shows detailed protocol messages.

### Manual Protocol Testing

Test protocol manually:

```bash
echo -e "GETINFO version\nBYE" | ./pinentry-proton
```

Expected output:
```
OK Proton Pass pinentry v1.0.0 ready
D 1.0.0
OK
OK closing connection
```

### Inspect Test Keys

GPG test key:

```bash
export GNUPGHOME=test/fixtures/gnupg
gpg --list-keys
```

SSH test key:

```bash
ssh-keygen -l -f test/fixtures/ssh/test_ed25519.pub
```

## Troubleshooting

### Tests Fail: "Binary not found"

Build first:
```bash
make build
```

### Tests Fail: "Test keys not found"

Create keys:
```bash
make test-setup
```

### GPG Test Fails: "No secret key"

Recreate GPG keys:
```bash
rm -rf test/fixtures/gnupg
./test/setup_test_keys.sh
```

### ProtonPass Integration Fails

Check:
1. ProtonPass CLI installed: `which pass`
2. Authenticated: `pass item list`
3. Test item exists with correct URI
4. Item contains password: `424242`

### SSH Test Fails on macOS

This is expected. macOS SSH prefers system keychain. The test includes fallback verification.

### Permission Errors

Ensure test scripts are executable:
```bash
chmod +x test/*.sh
```

## Security Notes

**⚠️ TEST KEYS ONLY**

- All test keys use passphrase `424242`
- Test keys are committed to the repository
- **NEVER use test keys in production**
- Test keys are for automated testing only

## Contributing Tests

When adding features, add tests:

1. **Unit tests** in the relevant `internal/*/` package
2. **Integration test** in `test/integration_test.go`
3. **Documentation** in this file

Run full test suite before submitting:

```bash
make check test-all
```

## Performance Testing

Benchmark protocol operations:

```bash
go test -bench=. -benchmem ./internal/protocol
```

## Test Maintenance

Regular maintenance tasks:

1. Update test keys annually
2. Test on new Go versions
3. Test on new macOS/Linux versions
4. Update ProtonPass CLI version
5. Review security best practices

## Summary

| Test Suite | Command | Duration | Requirements |
|------------|---------|----------|--------------|
| Unit | `make test` | ~2s | None |
| Unit (internal only) | `make test-unit` | ~1s | None |
| E2E (mock) | `make test-e2e` | ~3s | Binary built |
| Integration | `make test-integration` | ~5s | Binary built |
| Benchmarks | `make benchmark` | ~2min | None |
| Coverage Check | `make test-coverage-check` | ~2s | None |
| CI Suite | `make test-ci` | ~3s | None |
| Real ProtonPass | `make test-realpass` | varies | ProtonPass auth, test vault |
| GPG | `make test-gpg` | ~10s | GPG, test keys, ProtonPass |
| SSH | `make test-ssh` | ~5s | SSH tools, test keys, ProtonPass |
| All | `make test-all` | ~20s | All above |

**Coverage Metrics:**
- Overall: 82.4% (target: 75%)
- ProtonPass: 88.2%
- Platform: 100%
- Config: 87.1%
- Protocol: 77.4%

---

For questions or issues, see [CONTRIBUTING.md](CONTRIBUTING.md).
