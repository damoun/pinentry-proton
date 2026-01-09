# Project Architecture

This document describes the refactored architecture of pinentry-proton, following Go best practices for maintainability, reusability, and readability.

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
│   │   └── config_test.go         # Configuration tests
│   │
│   ├── protocol/
│   │   ├── session.go             # Pinentry protocol session management
│   │   ├── encoding.go            # Percent encoding/decoding utilities
│   │   ├── protocol_test.go       # Protocol unit tests
│   │   └── integration_test.go    # Protocol integration tests
│   │
│   ├── protonpass/
│   │   └── client.go              # ProtonPass CLI integration
│   │
│   └── platform/
│       ├── platform_darwin.go     # macOS-specific code
│       ├── platform_linux.go      # Linux-specific code
│       └── platform_other.go      # Unsupported platforms
│
├── docs/
│   ├── SECURITY.md                # Security policy
│   ├── CONTRIBUTING.md            # Contribution guidelines
│   └── CLAUDE.md                  # Claude Code guidance
│
├── config.example.yaml            # Example configuration
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build automation
├── README.md                      # User documentation
└── LICENSE                        # MIT License
```

## Package Overview

### `cmd/pinentry-proton`
**Purpose**: Application entry point

**Responsibilities**:
- Parse command-line arguments (if any)
- Load configuration
- Setup logging and debug mode
- Initialize platform-specific features
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

### `internal/platform`
**Purpose**: Platform-specific functionality

**Responsibilities**:
- Platform detection
- Platform-specific initialization
- Optional platform integration (Keychain, libsecret)
- Platform-specific cleanup

**Key Functions**:
- `Info()`: Get platform name
- `Setup()`: Platform initialization
- `Cleanup()`: Platform cleanup

**Build Tags**:
- `darwin`: macOS implementation
- `linux`: Linux implementation
- `!darwin,!linux`: Unsupported platforms

**Dependencies**: None

## Design Principles

### 1. **Separation of Concerns**
Each package has a single, well-defined responsibility:
- Configuration is isolated from protocol logic
- ProtonPass integration is separate from protocol
- Platform-specific code is isolated with build tags

### 2. **Dependency Management**
- `cmd/` depends on all `internal/` packages
- `internal/protocol` depends on `config` and `protonpass`
- `internal/config`, `internal/protonpass`, `internal/platform` have minimal dependencies
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
- Platform detection is isolated

### 5. **Maintainability**
- Clear package boundaries
- Each file has focused responsibility
- Tests are colocated with code
- Documentation at package level

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

### Platform Security
**Responsibility**: `internal/platform`
- Optional Keychain integration (macOS)
- Optional libsecret integration (Linux)
- Defaults to no persistent storage

## Data Flow

```
1. User Interaction (GPG/SSH agent)
         ↓
2. cmd/pinentry-proton/main.go
   - Load config (internal/config)
   - Setup platform (internal/platform)
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
   - Platform cleanup
```

## Testing Strategy

### Unit Tests
- **config**: Configuration loading, validation, matching
- **protocol**: Command parsing, encoding, session state

### Integration Tests
- **protocol**: Full protocol flows, edge cases, cancellation

### No Tests Required
- **cmd**: Entry point (integration tested via binary)
- **platform**: Stubs (no logic yet)
- **protonpass**: External dependency (tested via protocol tests)

## Build and Release

### Build
```bash
make build
# Builds: ./cmd/pinentry-proton → ./pinentry-proton
```

### Test
```bash
make test
# Tests: ./internal/config, ./internal/protocol
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
3. **Platform Integration**: Extend `internal/platform/platform_*.go`
4. **Additional Pass Providers**: New package in `internal/`

### Package Additions
Consider adding:
- `internal/cache/`: Optional password caching (encrypted)
- `internal/metrics/`: Optional telemetry (opt-in)
- `internal/gui/`: GUI pinentry mode
- `pkg/protocol/`: Export protocol for use by other tools

## Migration from Old Structure

### What Changed
- ✅ All Go code moved from root to packages
- ✅ Clear package boundaries established
- ✅ Tests colocated with code
- ✅ No code in repository root

### What Stayed the Same
- ✅ All functionality preserved
- ✅ API/behavior unchanged
- ✅ Configuration format unchanged
- ✅ Build process (Makefile) mostly unchanged
- ✅ All tests passing

### Benefits
- ✅ Better code organization
- ✅ Easier to understand
- ✅ Easier to test
- ✅ Easier to extend
- ✅ Follows Go best practices
- ✅ Enables future pkg/ exports

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
