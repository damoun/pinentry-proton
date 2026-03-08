# Integration Test Suite

This directory contains comprehensive integration tests for pinentry-proton.

## Quick Start

```bash
# Build the binary first
cd ..
make build

# Run all integration tests
make test-all
```

## Test Files

### `integration_test.go`

Go integration tests that test the pinentry protocol by running the binary and sending commands via stdin/stdout.

**Tests:**
- `TestBinaryExists` - Verify binary is built
- `TestBasicProtocol` - Basic Assuan protocol handshake
- `TestSetAndGetOptions` - SET* commands
- `TestGetPinWithMockProtonPass` - GETPIN with mock pass-cli
- `TestGetPinWithGPGContext` - GPG-specific context
- `TestCancelGetPin` - Error handling
- `TestInvalidCommands` - Invalid command handling
- `TestMultipleGetPin` - Multiple PINs in one session

**Run:**
```bash
go test -v ./test
# or short mode (skips ProtonPass tests):
go test -v -short ./test
```

### Shell Scripts

- `setup_test_keys.sh` - Creates GPG and SSH test keys
- `test_gpg.sh` - End-to-end GPG integration test
- `test_ssh.sh` - End-to-end SSH integration test
- `run_go_tests.sh` - Runs Go tests with race detector
- `run_all_tests.sh` - Master test runner

### `fixtures/`

Test data and configuration:
- `test-config.yaml` - Test configuration file
- `gnupg/` - GPG test keyring (created by setup_test_keys.sh)
- `ssh/` - SSH test keys (created by setup_test_keys.sh)

## Test Requirements

### Basic Tests

No requirements - these tests use mocks.

### GPG Integration Tests

Requires:
- GPG installed (`gpg` command)
- Test keys setup (`make test-setup`)
- ProtonPass CLI installed and authenticated
- Test item in ProtonPass with PIN 424242

### SSH Integration Tests

Requires:
- SSH tools (`ssh-agent`, `ssh-add`, `ssh-keygen`)
- Test keys setup (`make test-setup`)
- ProtonPass CLI installed and authenticated
- Test item in ProtonPass with PIN 424242

## ProtonPass Test Item

The tests expect a ProtonPass item with:
- **URI:** `pass://KmDQwh8YtmA3hKFn1x4KucB4ZXBG4_GXKLKp9oRP6uf_jn8wTTjzjnnP7A92KdQXmLp4kvgBAertdUZgggtZhQ==/MYhqRQ1mT5yo-l0TUh_Dzm38QvCsegOdKU2OWemXRheOOVAuv46qq7UBf6gWX3ZfiMDoOKnlfpSSPzAKRR_BRg==`
- **Password field:** `424242`

Create this item in your ProtonPass vault, or update the test configuration with your own test item.

## Running Specific Tests

```bash
# Unit tests only
make test

# Integration tests (Go)
make test-integration

# GPG tests
make test-gpg

# SSH tests
make test-ssh

# Everything
make test-all
```

## Debugging

Enable debug mode:
```bash
export PINENTRY_PROTON_DEBUG=1
./test_gpg.sh
```

Test protocol manually:
```bash
cd ..
echo -e "GETINFO version\nBYE" | ./pinentry-proton
```

## CI/CD

The Go integration tests can run in CI without ProtonPass:

```bash
make build
go test -short ./test
```

The `-short` flag skips tests that require ProtonPass.

For the full CI/CD configuration, see `../.github/workflows/ci.yml`

## Security

**⚠️ WARNING: Test keys are public!**

All test keys in this directory use the passphrase `424242` and are committed to the repository. **NEVER use these keys for anything other than testing!**

## Troubleshooting

### "Binary not found"
```bash
cd .. && make build
```

### "Test keys not found"
```bash
make test-setup
# or manually:
./setup_test_keys.sh
```

### GPG test fails
```bash
# Verify GPG installation
gpg --version

# Check test keys
export GNUPGHOME=$PWD/fixtures/gnupg
gpg --list-keys
```

### "Failed to retrieve password"
1. Check ProtonPass CLI: `pass item list`
2. Verify authentication: `pass auth status`
3. Check test item exists with correct URI
4. Verify item contains password "424242"

## Documentation

See [fixtures/README.md](fixtures/README.md) for test key documentation.

## Contributing

When adding tests:
1. Add unit tests in `internal/*/` packages
2. Add integration tests in `integration_test.go`
3. Update documentation
4. Ensure `make test-all` passes
