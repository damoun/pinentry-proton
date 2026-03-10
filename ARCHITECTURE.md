# Project Architecture

This document describes the architecture of pinentry-proton, following Go best practices for maintainability, reusability, and readability.

## Directory Structure

```
pinentry-proton/
├── cmd/
│   └── pinentry-proton/
│       └── main.go                 # Application entry point
│
├── internal/                       # Private application packages
│   ├── config/
│   │   ├── config.go              # Configuration loading and management
│   │   ├── config_test.go         # Configuration tests
│   │   └── config_benchmark_test.go
│   │
│   ├── protocol/
│   │   ├── session.go             # Pinentry protocol session management
│   │   ├── encoding.go            # Percent encoding/decoding utilities
│   │   ├── protocol_test.go       # Protocol unit tests
│   │   ├── protocol_benchmark_test.go
│   │   └── integration_test.go    # Protocol integration tests
│   │
│   └── protonpass/
│       ├── client.go              # ProtonPass CLI integration
│       ├── client_test.go         # ProtonPass client tests
│       └── client_benchmark_test.go
│
├── test/
│   ├── integration_test.go        # Binary protocol tests
│   ├── e2e/
│   │   └── e2e_test.go           # Full workflow tests with mock pass-cli
│   ├── testutil/
│   │   ├── fixtures.go           # Config helpers, assertions, setup
│   │   └── mock_pass.go          # Mock ProtonPass CLI implementation
│   └── fixtures/                  # Test configs, keys, mock data
│
├── config.example.yaml            # Example configuration
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build automation
├── README.md                      # User documentation
├── ARCHITECTURE.md                # This file
├── SECURITY.md                    # Security policy
├── CONTRIBUTING.md                # Contribution guidelines
├── CLAUDE.md                      # Claude Code guidance
└── LICENSE                        # MIT License
```

## Package Overview

### `cmd/pinentry-proton`
**Purpose**: Application entry point

**Responsibilities**:
- Load configuration
- Setup logging and debug mode
- Create and run protocol session
- Handle signals (SIGINT, SIGTERM)
- Coordinate cleanup on exit

**Dependencies**: All internal packages

### `internal/config`
**Purpose**: Configuration management

**Responsibilities**:
- Load configuration from multiple locations
- Parse YAML configuration files
- Validate configuration structure
- Match pinentry context to ProtonPass items
- Pattern matching for context criteria

**Key Types**:
- `Config`: Main configuration structure
- `Mapping`: Context-to-item mapping
- `MatchCriteria`: Matching rules for pinentry requests

**Key Functions**:
- `Load()`: Load configuration from standard locations
- `FindItemForContext()`: Match context to ProtonPass item
- `Validate()`: Validate configuration

**Dependencies**: `gopkg.in/yaml.v3`

### `internal/protocol`
**Purpose**: Pinentry Assuan protocol implementation

**Responsibilities**:
- Implement pinentry protocol commands
- Handle protocol session lifecycle
- Encode/decode percent-encoded data
- Track and cleanup sensitive data
- Integrate with ProtonPass client

**Key Types**:
- `Session`: Protocol session state

**Key Functions**:
- `NewSession()`: Create new protocol session
- `Run()`: Execute protocol session
- `UnescapeArg()`, `EscapeArg()`, `PercentEncode()`: Encoding utilities
- `Cleanup()`: Zero sensitive data

**Constants**:
- `Version`: Application version
- `DefaultTimeout`: Default operation timeout
- `DebugMode`: Debug logging flag

**Dependencies**:
- `internal/config`
- `internal/protonpass`

### `internal/protonpass`
**Purpose**: ProtonPass CLI integration

**Responsibilities**:
- Execute pass-cli commands
- Parse ProtonPass URIs
- Retrieve passwords from items
- Handle pass-cli errors
- Memory zeroing utilities

**Key Types**:
- `Client`: ProtonPass CLI client

**Key Functions**:
- `NewClient()`: Create new client
- `RetrievePassword()`: Get password from ProtonPass
- `ZeroBytes()`: Securely zero byte slices

**Dependencies**: None (standard library only)

## Design Principles

### 1. **Separation of Concerns**
Each package has a single, well-defined responsibility:
- Configuration is isolated from protocol logic
- ProtonPass integration is separate from protocol

### 2. **Dependency Management**
- `cmd/` depends on all `internal/` packages
- `internal/protocol` depends on `config` and `protonpass`
- `internal/config` and `internal/protonpass` have minimal dependencies
- Circular dependencies are impossible with `internal/` structure

### 3. **Testability**
- Each package can be tested independently
- Protocol tests don't require configuration loading
- Configuration tests don't require protocol execution
- Integration tests are separated from unit tests

### 4. **Reusability**
- Protocol encoding functions can be reused
- Configuration loading logic is independent
- ProtonPass client can be used standalone

### 5. **Maintainability**
- Clear package boundaries
- Each file has focused responsibility
- Tests are colocated with code

## Security Architecture

### Memory Safety
**Responsibility**: `internal/protonpass`
- `ZeroBytes()` function for secure memory clearing
- Used by all packages handling sensitive data

### Sensitive Data Tracking
**Responsibility**: `internal/protocol`
- `Session.sensitiveData` tracks all passwords
- `Cleanup()` ensures all tracked data is zeroed
- Deferred cleanup on signals and errors

### ProtonPass Integration
**Responsibility**: `internal/protonpass`
- No passwords in command-line arguments
- No passwords in environment variables
- Captures and clears stderr/stdout
- Timeouts prevent hanging

## Data Flow

```
1. User Interaction (GPG/SSH agent)
         ↓
2. cmd/pinentry-proton/main.go
   - Load config (internal/config)
   - Create session (internal/protocol)
         ↓
3. internal/protocol.Session.Run()
   - Parse pinentry commands
   - Match context (internal/config)
         ↓
4. internal/protonpass.Client.RetrievePassword()
   - Execute pass-cli
   - Parse output
   - Return password (tracked)
         ↓
5. internal/protocol.Session
   - Encode password
   - Send to agent
   - Zero password
         ↓
6. Cleanup
   - Zero all tracked passwords
```

## Testing Strategy

### Unit Tests
- **config**: Configuration loading, validation, matching
- **protocol**: Command parsing, encoding, session state
- **protonpass**: Client behavior, URI parsing, error handling

### Integration Tests
- **test/integration_test.go**: Full protocol flows via built binary
- **test/e2e/**: Complete workflows with mock pass-cli

### Benchmarks
- All internal packages include `*_benchmark_test.go` files
- Performance baselines tracked via `make benchmark-save`

## Build and Release

### Build
```bash
make build
# Builds: ./cmd/pinentry-proton → ./pinentry-proton
```

### Test
```bash
make test
# Tests: ./internal/...
```

### Install
```bash
make install
# Installs to: /usr/local/bin/pinentry-proton
```

## Future Enhancements

### Easy to Add
1. **New Protocol Commands**: Add to `internal/protocol/session.go`
2. **New Configuration Options**: Add to `internal/config/config.go`
3. **Additional Pass Providers**: New package in `internal/`

### Package Additions
Consider adding:
- `internal/cache/`: Optional password caching (encrypted)
- `internal/platform/`: Platform-specific integrations (Keychain, libsecret)
- `pkg/protocol/`: Export protocol for use by other tools

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
