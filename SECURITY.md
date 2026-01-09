# Pinentry-Proton Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, **do not** file a public issue. Instead, please email the maintainers at security@example.com with a detailed report including:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if available)

Security issues will be addressed promptly and coordinated disclosure will be provided.

## Threat Model

### Protected Against

✅ **Shoulder surfing** - Passwords retrieved automatically, no typing required  
✅ **Keyloggers** - No keyboard input of sensitive data  
✅ **Process listing exposure** - Passwords never in command-line arguments  
✅ **Memory persistence** - Sensitive data zeroed after use  
✅ **Signal interruption** - Proper cleanup on SIGINT/SIGTERM  

### Not Protected Against

❌ **Compromised ProtonPass account** - Attacker with account access can retrieve passwords  
❌ **System compromise with root access** - Root can access all process memory  
❌ **Memory dumps during use** - While password is in memory for transmission  
❌ **ProtonPass CLI vulnerabilities** - We depend on pass-cli security  
❌ **Man-in-the-middle on pass-cli execution** - Local process execution assumed secure

## Security Practices

This project follows security best practices for a pinentry tool:

### No Secret Logging

- **Never logs passwords, PINs, or passphrases** to any output stream
- No secrets in stderr, stdout, or files
- No telemetry or analytics that could leak secrets
- Test data uses mock values only

### Memory Management

- Sensitive data (passwords) stored in byte slices and **explicitly zeroed** after use
- String conversions are minimized and avoided where possible
- Memory cleared on normal exit and signal interruption
- Uses `zeroBytes()` function to overwrite sensitive buffers

### Input Isolation

- Passwords **never** accepted via command-line arguments
- Passwords **never** stored in environment variables
- No persistent storage of passwords
- Retrieved on-demand from ProtonPass via pass-cli

### Protocol Implementation

- Implements Assuan protocol correctly
- Proper percent-encoding/decoding of data
- Validates all inputs
- Returns appropriate error codes

### Signal Handling

- Gracefully handles SIGINT and SIGTERM
- Ensures cleanup before exit
- Context-based cancellation for all operations

### Timeout Protection

- Configurable timeout for password retrieval (default: 60s)
- Prevents hanging on network issues
- Context cancellation on timeout

### Dependencies

- **Minimal dependencies**: Only golang.org/x/term and gopkg.in/yaml.v3
- Standard library preferred
- Dependencies are pinned in go.mod
- Regular security audits

### Platform Security

**macOS:**
- Follows macOS security guidelines
- Code signing and notarization recommended for distribution
- Uses standard ProtonPass CLI

**Linux:**
- No special privileges required
- Works with standard user permissions
- Uses standard ProtonPass CLI

## Configuration Security

Your configuration file contains:
- Vault names or Share IDs
- Item titles or Item IDs  
- Mapping logic

**Recommendations:**

✅ Use `chmod 600 ~/.config/pinentry-proton/config.yaml`  
✅ Use Share IDs and Item IDs instead of names when possible (less info disclosure)  
✅ Review home directory permissions  
✅ Don't commit config.yaml to version control  
✅ Use config.example.yaml as template only  

**Note:** The configuration file does NOT contain passwords, only references to where they're stored.

## Dependency Management

- Dependencies are pinned to specific versions in `go.mod` and `go.sum`
- Regular security updates applied
- Dependency scanning performed in CI/CD via Gosec
- Supply chain attacks mitigated by go.sum verification

## CI/CD Security

- **No secrets** are stored in CI/CD pipelines
- Test data uses mock/ephemeral values only
- Security scanning with Gosec
- Build artifacts checksummed for integrity
- No ProtonPass credentials in tests

## Auditing

### Code Review Checklist

When reviewing code changes, verify:

- [ ] No secrets in code, tests, or comments
- [ ] Memory zeroing after password use
- [ ] Proper signal handling and cleanup
- [ ] No logging of sensitive data
- [ ] Dependencies are vetted and minimal
- [ ] Tests don't require real ProtonPass credentials
- [ ] Error messages don't leak secrets

### Static Analysis

Run these checks:

```bash
make vet          # Go vet
make lint         # golangci-lint with gosec
go mod verify     # Verify dependencies
```

## Known Limitations

### Architectural Limitations

- Passwords must exist in memory briefly during transmission to agent
- Depends on ProtonPass CLI security model
- Requires local pass-cli installation
- Configuration file discloses vault/item structure (but not passwords)

### Environmental Requirements

- User must have active ProtonPass session
- User must have read access to configured items
- pass-cli must be in PATH
- Network access required for ProtonPass API

### Not Designed For

❌ Headless/non-interactive environments without configuration  
❌ Multi-user systems where users share home directories  
❌ Systems where root user is untrusted  
❌ Air-gapped systems (requires ProtonPass API access)  

## Compliance and Standards

This project aims to follow:

- **OWASP Secure Coding Practices** - Input validation, output encoding, secrets management
- **CWE/SANS Top 25** - Mitigations for common weaknesses
- **Go Security Best Practices** - Memory safety, dependency management
- **Pinentry Protocol Specification** - Correct Assuan protocol implementation

## Security Updates

When a security issue is found:

1. **Disclosure** - Private notification to maintainers
2. **Assessment** - Severity and impact evaluation (24-48 hours)
3. **Fix** - Develop and test patch
4. **Release** - New version with security fix
5. **Advisory** - Public disclosure with CVE if applicable
6. **Notification** - Users notified via GitHub releases

## Questions?

For security questions or concerns that aren't vulnerabilities, open a GitHub discussion or issue.
