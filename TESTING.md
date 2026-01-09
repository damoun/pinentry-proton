# Testing Guide

This document describes the comprehensive test suite for pinentry-proton.

## Overview

A comprehensive test suite covering unit tests, protocol integration, and real-world GPG/SSH workflows. All tests use PIN **424242** and the ProtonPass test item URI (see Test Configuration below).

### Test Suite Summary

**Test Types:**
- **Unit Tests** - Go package tests for config, protocol, and protonpass packages
- **Integration Tests** - Binary protocol tests using real pinentry commands
- **GPG Integration** - Real GPG signing, verification, encryption, and decryption
- **SSH Integration** - Real SSH key operations and agent interaction

**Test Infrastructure:**
- 8 Go integration tests covering complete pinentry protocol
- GPG end-to-end test with real cryptographic operations
- SSH end-to-end test with real key operations
- Mock ProtonPass CLI for isolated testing
- Test key infrastructure (GPG + SSH with passphrase 424242)
- Comprehensive Makefile targets for easy test execution

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
├── fixtures/              # Test data and keys
│   ├── gnupg/            # GPG test keyring
│   ├── ssh/              # SSH test keys
│   ├── test-config.yaml  # Test configuration
│   └── README.md         # Fixture documentation
├── integration_test.go   # Go integration tests
├── setup_test_keys.sh    # Key creation script
├── test_gpg.sh          # GPG integration test
├── test_ssh.sh          # SSH integration test
├── run_go_tests.sh      # Go test runner
└── run_all_tests.sh     # Master test runner
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

The test suite is designed for CI environments:

```yaml
# Example GitHub Actions
- name: Setup test keys
  run: make test-setup

- name: Run all tests
  run: make test-all
  env:
    PINENTRY_PROTON_CONFIG: test/fixtures/test-config.yaml
```

**Important:** Never commit real ProtonPass credentials to CI.

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
| Integration | `make test-integration` | ~5s | Binary built |
| GPG | `make test-gpg` | ~10s | GPG, test keys, ProtonPass |
| SSH | `make test-ssh` | ~5s | SSH tools, test keys, ProtonPass |
| All | `make test-all` | ~20s | All above |

---

For questions or issues, see [CONTRIBUTING.md](CONTRIBUTING.md).
