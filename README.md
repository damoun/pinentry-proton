# Pinentry-Proton

A secure pinentry program that integrates ProtonPass with GPG and SSH agents. This tool retrieves passwords from your ProtonPass vaults to unlock SSH keys and GPG keys, eliminating the need to manually type passphrases while maintaining security.

## Table of Contents

- [Quick Start](#quick-start)
- [Features](#features)
- [How It Works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [ProtonPass Setup](#protonpass-setup)
- [Usage](#usage)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Secure**: Follows pinentry protocol security best practices
- **Memory safety**: Zeros sensitive data after use
- **Signal handling**: Properly handles SIGINT and SIGTERM
- **Configurable**: Flexible mapping between keys and ProtonPass items
- **Cross-platform**: Works on macOS and Linux
- **No secrets in logs**: Never logs passwords or sensitive data
- **Timeout support**: Configurable timeouts for password retrieval

## Quick Start

Complete setup in 5 steps:

### 1. Install prerequisites

```bash
# macOS
brew install gnupg protonpass-cli

# Linux (Debian/Ubuntu)
apt install gnupg
# Install pass-cli from https://github.com/protonpass/pass-cli/releases
```

### 2. Install pinentry-proton

```bash
git clone https://github.com/damoun/pinentry-proton.git
cd pinentry-proton
make build
sudo make install
```

### 3. Store your passphrase in ProtonPass

```bash
# Log in to ProtonPass CLI
pass-cli login

# Store your GPG or SSH key passphrase
pass-cli item create login \
  --vault "Personal" \
  --title "GPG Key" \
  --password "your-key-passphrase"
```

### 4. Create the configuration file

```bash
mkdir -p ~/.config/pinentry-proton
cat > ~/.config/pinentry-proton/config.yaml << 'EOF'
default_item: "pass://Personal/GPG Key/password"
EOF
```

For multiple keys, see [Configuration](#configuration) for mapping rules.

### 5. Configure GPG agent to use pinentry-proton

```bash
mkdir -p ~/.gnupg
echo "pinentry-program /usr/local/bin/pinentry-proton" >> ~/.gnupg/gpg-agent.conf
gpgconf --kill gpg-agent
```

That's it. The next GPG or SSH operation will retrieve your passphrase from ProtonPass automatically.

---

### YubiKey / Smartcard Setup

YubiKey GPG cards prompt for a PIN when signing. To retrieve the PIN from ProtonPass:

**1. Find your card's key grip:**
```bash
gpg --card-status
# Note the "General key info" fingerprint
gpg --with-keygrip -K YOUR_KEY_FINGERPRINT
# Note the "Keygrip" value
```

**2. Store your card PIN in ProtonPass:**
```bash
pass-cli item create login \
  --vault "Personal" \
  --title "YubiKey PIN" \
  --password "your-yubikey-pin"
```

**3. Map the keygrip in your config:**
```yaml
mappings:
  - name: "YubiKey PIN"
    item: "pass://Personal/YubiKey PIN/password"
    match:
      keyinfo: "YOUR_KEYGRIP_HERE"
```

**Tip:** If you don't know the exact keygrip yet, use `PINENTRY_PROTON_DEBUG=1` while running a GPG operation to see the `keyinfo` value in the logs, then add it to your config.

---

## Workflow: Sign Commits Without Repeated PIN Entry

**Primary Use Case:** Sign git commits (and perform other GPG/SSH operations) seamlessly without manual PIN entry.

**How it works in practice:**

1. **Unlock ProtonPass once** - Authenticate to ProtonPass UI/CLI at the start of your session
2. **Work normally** - When GPG needs to unlock your smartcard/key (e.g., to sign a commit), it calls pinentry-proton
3. **Automatic retrieval** - pinentry-proton automatically fetches your PIN from ProtonPass
4. **Seamless operation** - Your commit gets signed without interrupting your workflow
5. **Multiple operations** - Sign multiple commits, encrypt files, SSH operations - all without re-entering PINs

**Example workflow:**
```bash
# Unlock ProtonPass once at the start of your day
pass-cli login

# Now sign commits seamlessly
git commit -S -m "feat: add new feature"    # Signs without prompting for PIN
git commit -S -m "fix: resolve bug"         # Signs without prompting for PIN
git commit -S -m "docs: update README"      # Signs without prompting for PIN

# SSH operations also work seamlessly
ssh git@github.com                           # Uses key without prompting for passphrase
```

**Key Benefit:** Eliminates repetitive PIN/passphrase entry while maintaining security through ProtonPass authentication.

## How It Works

1. GPG/SSH agent requests a PIN via the pinentry protocol
2. Pinentry-Proton matches the request to a configured ProtonPass item
3. Retrieves the password using `pass-cli item view`
4. Securely returns the password to the agent
5. Zeros password from memory

## Prerequisites

- Go 1.21 or later (for building)
- [ProtonPass CLI (pass-cli)](https://github.com/protonpass/pass-cli) installed and configured
- Active ProtonPass session (`pass-cli login`)
- SSH keys or GPG keys stored in ProtonPass with their passphrases

## Installation

### From Source

```bash
git clone https://github.com/damoun/pinentry-proton.git
cd pinentry-proton
make build
sudo make install  # Installs to /usr/local/bin
```

### Manual Installation

```bash
go build -o pinentry-proton
sudo install -m 755 pinentry-proton /usr/local/bin/
```

## Configuration

### 1. Create Configuration File

Create `~/.config/pinentry-proton/config.yaml`:

```yaml
# Default item if no mapping matches (optional)
default_item: "pass://Personal/Default SSH Key/password"

# Timeout in seconds (default: 60)
timeout: 60

# Map pinentry contexts to ProtonPass items
mappings:
  - name: "GitHub SSH Key"
    item: "pass://Work/GitHub SSH Key/password"
    match:
      description: "github"

  - name: "Personal GPG Key"
    item: "pass://Personal/GPG Key/passphrase"
    match:
      keyinfo: "ABCD1234"  # GPG key ID
```

See [config.example.yaml](config.example.yaml) for more examples.

### 2. Configure GPG Agent

Edit `~/.gnupg/gpg-agent.conf`:

```conf
pinentry-program /usr/local/bin/pinentry-proton
```

Reload the agent:

```bash
gpgconf --kill gpg-agent
```

### 3. Configure SSH Agent

For SSH keys with passphrases, configure your SSH environment to use the pinentry:

#### macOS

Edit `~/.ssh/config`:

```
Host *
    UseKeychain no
```

Then ensure your SSH agent uses GPG agent:

```bash
export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)
```

Add this to your `~/.zshrc` or `~/.bashrc`.

#### Linux

Configure GPG agent to handle SSH:

Edit `~/.gnupg/gpg-agent.conf`:

```conf
enable-ssh-support
pinentry-program /usr/local/bin/pinentry-proton
```

Add to your shell profile:

```bash
export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)
gpgconf --launch gpg-agent
```

## ProtonPass Setup

### Store SSH Key Passphrases

1. Create or import your SSH key in ProtonPass:
   - If creating new: Use pass-cli to generate it
   - If importing: Store the passphrase in a Login or Note item

2. Note the vault and item details for configuration

### Example: Import Existing SSH Key

If you have an SSH key with a passphrase:

```bash
# Create an item in ProtonPass with the passphrase
pass-cli item create login \
  --share-id "YOUR_VAULT_SHARE_ID" \
  --title "GitHub SSH Key" \
  --password "your-ssh-key-passphrase"

# Configure pinentry-proton to use it
# In config.yaml:
# - name: "GitHub SSH Key"
#   item: "pass://YOUR_VAULT/GitHub SSH Key/password"
#   match:
#     description: "github"
```

### Finding Your Item URI

Use `pass-cli item view` to look up an item and confirm the exact vault and title to use in your URI:

```bash
# Look up by vault name and item title
pass-cli item view --vault-name "Personal" --item-title "GPG Key"

# Once you have the correct names, verify the URI resolves correctly
pass-cli item view 'pass://Personal/GPG Key/password'
```

If the second command returns your passphrase, the URI is correct and ready to use in `config.yaml`.

### ProtonPass URI Format

Use one of these formats:

```
pass://SHARE_ID/ITEM_ID/password
pass://VAULT_NAME/ITEM_TITLE/password
```

Examples:
- `pass://abc123def/xyz789/password` (by IDs)
- `pass://Work/GitHub SSH Key/password` (by names)
- `pass://Personal/GPG Key/passphrase` (custom field)

## Usage

### SSH Key Unlocking

When you use an SSH key:

```bash
ssh user@server
# pinentry-proton automatically retrieves the passphrase from ProtonPass
```

### GPG Key Unlocking

When GPG needs your key:

```bash
gpg --sign document.txt
# pinentry-proton automatically retrieves the passphrase from ProtonPass
```

### Testing

Test the pinentry directly:

```bash
echo -e "SETDESC Test description\nGETPIN\nBYE" | pinentry-proton
```

Expected output:
```
OK Proton Pass pinentry v1.0.0 ready
OK
D <encoded-password>
OK
OK
```

## Configuration Matching

Pinentry-Proton matches requests using these fields from GPG/SSH agents:

- **description**: Set by `SETDESC` (often contains key purpose)
- **prompt**: Set by `SETPROMPT` (usually "PIN:" or "Passphrase:")
- **title**: Set by `SETTITLE` (dialog title)
- **keyinfo**: Set by `SETKEYINFO` (GPG key ID or SSH key fingerprint)

### Matching Rules

- Matching is case-insensitive substring matching
- All specified criteria must match (AND logic)
- First matching mapping wins
- If no mapping matches, `default_item` is used
- If no `default_item` and no match, returns an error

### Finding the Right Match

To discover what values to match on:

1. Set `PINENTRY_PROTON_DEBUG=1` environment variable
2. Trigger a PIN request (SSH or GPG operation)
3. Check the debug output to see what values were received
4. Update your configuration accordingly

## Security Considerations

### What This Tool Does

✅ Retrieves passwords from ProtonPass securely
✅ Zeros passwords from memory after use
✅ Never logs passwords or sensitive data
✅ Handles signals gracefully (SIGINT, SIGTERM)
✅ Uses secure ProtonPass CLI communication
✅ Implements proper pinentry protocol

### What This Tool Does NOT Do

❌ Does not store passwords persistently
❌ Does not cache passwords in memory
❌ Does not expose passwords via command-line arguments
❌ Does not write passwords to disk or logs

### Prerequisites for Security

You must:

- Keep your ProtonPass account secure
- Use a strong master password
- Keep your ProtonPass session secure
- Protect your configuration file (contains vault/item names)
- Review the ProtonPass CLI security model

### Threat Model

**Protected against:**
- Shoulder surfing (no typing passphrases)
- Keyloggers (passwords not typed)
- Process listing (no passwords in argv)

**Not protected against:**
- Compromised ProtonPass account
- Compromised system with root access
- Memory dumps while password is in use
- ProtonPass CLI vulnerabilities

### Configuration Security

Your `config.yaml` contains:
- Vault names
- Item titles
- Mapping logic

**Recommendations:**
- Use `chmod 600 ~/.config/pinentry-proton/config.yaml`
- Use Share IDs and Item IDs instead of names when possible
- Review access to your home directory

## Troubleshooting

### "No ProtonPass item configured for this context"

**Solution:** Add a mapping or default_item to your config.yaml

### "Failed to retrieve password"

**Possible causes:**
- Not logged into ProtonPass: `pass-cli login`
- Incorrect item URI in configuration
- ProtonPass item not accessible
- Network issues

**Debug:**
```bash
# Find the exact vault and item names
pass-cli item view --vault-name "YOUR_VAULT" --item-title "YOUR_ITEM"

# Verify the URI resolves correctly
pass-cli item view 'pass://YOUR_VAULT/YOUR_ITEM/password'
```

### "pass-cli: command not found"

**Solution:** Install ProtonPass CLI:
```bash
# macOS
brew install protonpass-cli

# Linux - see ProtonPass CLI documentation
```

### SSH Keys Not Using Pinentry

**Check:**
1. Is GPG agent running with SSH support?
   ```bash
   echo $SSH_AUTH_SOCK
   # Should point to GPG agent socket
   ```

2. Is pinentry-program configured?
   ```bash
   grep pinentry-program ~/.gnupg/gpg-agent.conf
   ```

3. Reload agent:
   ```bash
   gpgconf --kill gpg-agent
   ```

### GPG Not Using Pinentry

**Check:**
```bash
gpgconf --list-dirs
# Verify configuration directory

cat ~/.gnupg/gpg-agent.conf
# Verify pinentry-program is set

gpgconf --kill gpg-agent
# Restart agent
```

## Development

### Building

```bash
make build           # Build the binary
make install         # Install to /usr/local/bin
make lint            # Run linters
make coverage        # Generate test coverage report
```

### Pre-Commit Hooks

This project uses [pre-commit](https://pre-commit.com/) hooks to ensure code quality. Hooks are split between commit and push stages for optimal developer experience:

#### Installation

```bash
# Install pre-commit tool (if not already installed)
pip install pre-commit
# or
brew install pre-commit

# Install hooks for this repository
make pre-commit-install
```

#### Hook Stages

**Commit Stage (~1-2s)** - Fast checks on every `git commit`:
- `go fmt` - Code formatting
- `go vet` - Static analysis
- `go test` - Unit tests (without race detection)
- `go mod verify` - Module integrity
- Branch protection - Prevent commits to main

**Push Stage (~15-25s)** - Comprehensive checks on every `git push`:
- All commit-stage checks (redundant safety net)
- `go test -race` - Unit tests with race detection
- `golangci-lint` - Full linting (11 linters)
- `go mod tidy` - Verify modules are tidy
- `make build` - Build verification
- Secrets scan - Check for accidentally committed secrets

#### Manual Execution

Run hooks manually without committing/pushing:

```bash
make pre-commit-run-commit    # Run commit-stage hooks on all files
make pre-commit-run-push      # Run push-stage hooks on all files
make pre-commit-run-all       # Run all hooks on all files
```

Or use pre-commit directly:

```bash
pre-commit run --hook-stage commit --all-files
pre-commit run --hook-stage push --all-files
```

#### Bypassing Hooks

In rare cases when you need to skip hooks:

```bash
git commit --no-verify    # Skip commit hooks
git push --no-verify      # Skip push hooks
```

**Note:** Use sparingly. Hooks exist to catch issues before they reach CI/CD.

#### CI/CD Parity

The push-stage hooks mirror the CI/CD pipeline, ensuring that if push hooks pass, CI should pass (barring environment differences).

### Testing

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

**Quick Start:**

```bash
make test-setup      # Create test keys (first time only)
make test            # Run unit tests
make test-all        # Run all tests (unit + integration + GPG + SSH)
```

For detailed testing information, test requirements, and troubleshooting, see:
- [TESTING.md](TESTING.md) - Comprehensive testing guide
- [TEST_QUICKREF.md](TEST_QUICKREF.md) - Quick reference card

## Project Structure

```
pinentry-proton/
├── cmd/pinentry-proton/       # Application entry point
├── internal/
│   ├── config/                # Configuration management
│   ├── protocol/              # Pinentry protocol implementation
│   ├── protonpass/            # ProtonPass CLI integration
│   └── platform/              # Platform-specific code (macOS/Linux)
├── test/                      # Test suite and fixtures
├── config.example.yaml        # Example configuration
└── Makefile                   # Build automation
```

**Documentation:**
- [ARCHITECTURE.md](ARCHITECTURE.md) - Detailed technical architecture
- [SECURITY.md](SECURITY.md) - Security policy and threat model
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [CLAUDE.md](CLAUDE.md) - Claude Code guidance

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Security requirements and best practices
- Code standards and testing requirements
- Pull request process
- Development workflow

## Pinentry Protocol Reference

This implementation follows the [Assuan protocol](https://www.gnupg.org/documentation/manuals/assuan/) used by GnuPG.

Key commands supported:
- `SETDESC`: Set description text
- `SETPROMPT`: Set prompt text
- `SETTITLE`: Set dialog title
- `SETERROR`: Set error message
- `SETKEYINFO`: Set key information
- `GETPIN`: Request PIN/passphrase
- `GETINFO`: Get information about pinentry
- `BYE`: End session

## Related Projects

- [ProtonPass CLI](https://github.com/protonpass/pass-cli) - Official ProtonPass command-line interface
- [GnuPG](https://gnupg.org/) - GNU Privacy Guard
- [pinentry](https://www.gnupg.org/related_software/pinentry/) - Collection of simple PIN or passphrase entry dialogs

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Security

See [SECURITY.md](SECURITY.md) for our security policy and how to report vulnerabilities.
