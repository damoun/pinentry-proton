# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pinentry-Proton is a secure pinentry program that integrates ProtonPass with GPG and SSH agents. It implements the Assuan pinentry protocol to retrieve passwords from ProtonPass vaults using the official ProtonPass CLI (`pass-cli`), eliminating the need to manually type passphrases while maintaining security.

**Core Purpose**: Act as a pinentry replacement that retrieves passwords from ProtonPass instead of prompting the user interactively.

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
# Run unit tests only
make test

# Run unit tests with race detection
go test -race ./...

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

# Generate coverage report
make coverage
```

### Running Single Tests
```bash
# Test a specific package
go test -v ./internal/config
go test -v ./internal/protocol

# Test a specific function
go test -v -run TestSpecificFunction ./internal/config
```

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

### Test Requirements
- **Test PIN**: All tests use `424242` as the test passphrase
- **Test ProtonPass Item**: Use the URI in `test/fixtures/test-config.yaml`
- **Mock pass-cli**: Integration tests create a mock `pass` binary returning `424242`

### Test Structure
```
test/fixtures/           Test data, keys, mock configs
test/integration_test.go Go integration tests (binary + protocol)
test/test_gpg.sh         End-to-end GPG workflow
test/test_ssh.sh         End-to-end SSH workflow
test/run_all_tests.sh    Master test runner
```

### Test Keys
- Located in `test/fixtures/gnupg/` (GPG) and `test/fixtures/ssh/` (SSH)
- All use passphrase `424242`
- **Never use in production** - committed for testing only

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
