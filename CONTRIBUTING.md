# Contributing to Pinentry-Proton

Thank you for your interest in contributing to pinentry-proton! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project follows standard open-source community guidelines. Be respectful, constructive, and professional in all interactions.

## Security First

This is a security-sensitive project that handles passwords and authentication credentials. All contributions must prioritize security:

- **Review [SECURITY.md](SECURITY.md)** for the threat model and security policy
- **See [CLAUDE.md](CLAUDE.md)** for critical security requirements and implementation patterns
- **Never commit secrets** to the repository
- **Zero sensitive data** from memory after use
- **No logging** of passwords, PINs, or passphrases
- **Follow Go security best practices** for memory management

## Getting Started

### Prerequisites

- Go 1.21 or later
- ProtonPass CLI (`pass-cli`) for integration testing
- golangci-lint for linting
- git

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/damoun/pinentry-proton.git
cd pinentry-proton

# Install dependencies
go mod download

# Verify dependencies
go mod verify

# Run tests
make test

# Run linters
make lint
```

## Making Changes

### 1. Create an Issue

Before starting work on a new feature or significant change:

1. Check existing issues to avoid duplication
2. Create a new issue describing the change
3. Wait for discussion and approval for significant changes

For bug fixes and small improvements, you can proceed directly to a pull request.

### 2. Branch Strategy

- Create a feature branch from `main`
- Use descriptive branch names: `feature/add-retry-logic`, `fix/memory-leak`, `docs/update-readme`

```bash
git checkout -b feature/your-feature-name
```

### 3. Code Standards

#### Go Code Style

- Follow standard Go formatting: `go fmt ./...`
- Follow Go naming conventions
- Write idiomatic Go code
- Keep functions focused and small
- Add comments for exported functions and types
- Use meaningful variable names

#### Security Requirements

**CRITICAL:** All code must pass security review:

- ✅ No secrets in code, comments, or tests
- ✅ Memory zeroing for sensitive data
- ✅ Proper error handling without leaking info
- ✅ Input validation and sanitization
- ✅ Context cancellation support
- ✅ Signal handling and cleanup
- ✅ Minimal dependencies

#### Code Example

```go
// Good: Secure password handling
func handlePassword(ctx context.Context, pass []byte) error {
    defer zeroBytes(pass)  // Always zero after use
    
    // Process password
    if err := process(pass); err != nil {
        return fmt.Errorf("processing failed: %w", err)  // Don't leak password
    }
    return nil
}

// Bad: Password leaked in error
func badHandlePassword(pass []byte) error {
    if err := process(pass); err != nil {
        return fmt.Errorf("failed with password %s: %w", string(pass), err)  // NEVER!
    }
    return nil
}
```

### 4. Testing

All changes must include appropriate tests:

- **Unit tests** for new functions
- **Integration tests** for protocol flows
- **Security tests** for memory zeroing and cleanup
- **Error cases** must be tested

```bash
# Run tests
make test

# Run tests with coverage
make coverage

# Run race detector
go test -race ./...
```

#### Test Guidelines

- Use table-driven tests where appropriate
- Test both success and failure cases
- Never use real ProtonPass credentials in tests
- Use mock data and temporary files
- Clean up test artifacts

### 5. Documentation

Update documentation for all user-facing changes:

- Update `README.md` for new features or usage changes
- Update `config.example.yaml` for configuration changes
- Add inline code comments for complex logic
- Update `SECURITY.md` for security-related changes

### 6. Commit Messages

Write clear, descriptive commit messages:

```
feat: add retry logic for pass-cli failures

- Implement exponential backoff
- Add max retry configuration option
- Update tests for retry scenarios

Fixes #123
```

Format:
- **feat:** New feature
- **fix:** Bug fix
- **docs:** Documentation only
- **test:** Adding or updating tests
- **refactor:** Code refactoring
- **perf:** Performance improvement
- **security:** Security improvement
- **chore:** Maintenance tasks

### 7. Pull Request Process

1. **Before submitting:**
   ```bash
   # Format code
   make fmt
   
   # Run linters
   make lint
   
   # Run all tests
   make test
   
   # Build successfully
   make build
   ```

2. **Create pull request:**
   - Use a descriptive title
   - Reference related issues
   - Describe what changed and why
   - Include testing steps
   - Note any breaking changes

3. **PR template:**
   ```markdown
   ## Description
   Brief description of the change
   
   ## Motivation
   Why is this change needed?
   
   ## Changes
   - List of changes made
   
   ## Testing
   - How was this tested?
   - Are there new tests?
   
   ## Security Impact
   - Any security implications?
   - Security checklist completed?
   
   ## Checklist
   - [ ] Tests pass
   - [ ] Linters pass
   - [ ] Documentation updated
   - [ ] No secrets in code
   - [ ] Memory zeroing implemented
   - [ ] Reviewed SECURITY.md and CLAUDE.md
   ```

4. **Review process:**
   - Maintainers will review your PR
   - Address feedback and requested changes
   - Keep PR scope focused
   - Be responsive to comments

## Project Structure

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed project structure and package organization.

## Adding New Features

### Feature Checklist

Before adding a new feature:

- [ ] Is it aligned with the project's purpose?
- [ ] Does it maintain security standards?
- [ ] Is it cross-platform (macOS & Linux)?
- [ ] Can it be tested without real credentials?
- [ ] Does it add minimal dependencies?
- [ ] Is the API/config backward compatible?

### Breaking Changes

Breaking changes require:

1. Discussion in an issue first
2. Major version bump
3. Migration guide in release notes
4. Update to README with migration instructions

## Dependencies

### Adding Dependencies

Minimize dependencies. When adding a new dependency:

1. Justify why it's needed
2. Verify it's well-maintained
3. Check for known vulnerabilities
4. Prefer standard library
5. Pin to specific versions

```bash
# Add dependency
go get package@version

# Verify and tidy
go mod verify
go mod tidy
```

### Allowed Dependencies

Current allowed dependencies:
- Standard library (preferred)
- `golang.org/x/term` (for secure terminal input)
- `gopkg.in/yaml.v3` (for config parsing)

## Code Review

### What Reviewers Look For

- **Security:** No leaked secrets, proper memory handling
- **Correctness:** Code does what it claims
- **Tests:** Adequate test coverage
- **Style:** Follows Go conventions
- **Documentation:** Changes are documented
- **Simplicity:** Code is as simple as possible

### As a Reviewer

- Be constructive and respectful
- Focus on the code, not the person
- Explain the "why" behind suggestions
- Approve when standards are met

## Release Process

Releases are managed by maintainers:

1. Version bumps follow [Semantic Versioning](https://semver.org/)
2. Changelog updated for each release
3. Binaries built for macOS and Linux
4. Checksums generated for verification
5. Release notes include migration guides if needed

## Questions?

- Open an issue for general questions
- Check existing issues and documentation first
- For security issues, see [SECURITY.md](SECURITY.md)

## Recognition

Contributors will be acknowledged in release notes and the README. Thank you for making pinentry-proton better!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
