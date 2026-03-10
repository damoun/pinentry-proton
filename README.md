# Pinentry-Proton

A secure pinentry program that retrieves passwords from [ProtonPass](https://github.com/protonpass/pass-cli) for GPG and SSH operations, eliminating manual passphrase entry.

## Quick Start

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
pass-cli login
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

### 5. Configure GPG agent

```bash
mkdir -p ~/.gnupg
echo "pinentry-program /usr/local/bin/pinentry-proton" >> ~/.gnupg/gpg-agent.conf
gpgconf --kill gpg-agent
```

That's it. GPG and SSH operations will now retrieve passphrases from ProtonPass automatically.

## How It Works

1. GPG/SSH agent requests a PIN via the pinentry protocol
2. Pinentry-Proton matches the request to a configured ProtonPass item
3. Retrieves the password using `pass-cli item view pass://VAULT/ITEM/FIELD`
4. Returns the password to the agent and zeros it from memory

**Workflow:** Unlock ProtonPass once at the start of your session, then sign commits, encrypt files, and SSH without re-entering PINs.

## Prerequisites

- Go 1.21+ (for building)
- [ProtonPass CLI (pass-cli)](https://github.com/protonpass/pass-cli) installed and configured
- Active ProtonPass session (`pass-cli login`)

## Configuration

Create `~/.config/pinentry-proton/config.yaml`:

```yaml
# Default item if no mapping matches
default_item: "pass://Personal/Default Key/password"

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
      keyinfo: "ABCD1234"
```

See [config.example.yaml](config.example.yaml) for more examples.

Config file locations (checked in order):
1. `$PINENTRY_PROTON_CONFIG`
2. `$XDG_CONFIG_HOME/pinentry-proton/config.yaml`
3. `$HOME/.config/pinentry-proton/config.yaml`
4. `$HOME/.pinentry-proton.yaml`

### Matching Rules

- Case-insensitive substring matching
- All specified criteria must match (AND logic)
- First matching mapping wins
- Falls back to `default_item` if no match

Use `PINENTRY_PROTON_DEBUG=1` to see what context values GPG/SSH sends, then configure your mappings accordingly.

### ProtonPass URI Format

```
pass://VAULT_NAME/ITEM_TITLE/FIELD
pass://SHARE_ID/ITEM_ID/FIELD
```

Field defaults to `password` if omitted. Verify your URI works with:

```bash
pass-cli item view 'pass://Personal/GPG Key/password'
```

### YubiKey / Smartcard Setup

YubiKey GPG cards prompt for a PIN when signing. To automate this:

1. Find your card's keygrip: `gpg --card-status` then `gpg --with-keygrip -K YOUR_FINGERPRINT`
2. Store your card PIN in ProtonPass
3. Map the keygrip in your config:

```yaml
mappings:
  - name: "YubiKey PIN"
    item: "pass://Personal/YubiKey PIN/password"
    match:
      keyinfo: "YOUR_KEYGRIP_HERE"
```

### SSH Agent Setup

Configure GPG agent to handle SSH by adding to `~/.gnupg/gpg-agent.conf`:

```conf
enable-ssh-support
pinentry-program /usr/local/bin/pinentry-proton
```

Then add to your shell profile:

```bash
export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket)
gpgconf --launch gpg-agent
```

On macOS, also add `UseKeychain no` to `~/.ssh/config`.

## Troubleshooting

**"No ProtonPass item configured for this context"** — Add a mapping or `default_item` to your config.

**"Failed to retrieve password"** — Check: `pass-cli login`, correct item URI, network connectivity. Debug with:
```bash
pass-cli item view 'pass://YOUR_VAULT/YOUR_ITEM/password'
```

**"pass-cli: command not found"** — Install ProtonPass CLI from [protonpass/pass-cli](https://github.com/protonpass/pass-cli).

**Agent not using pinentry** — Verify `pinentry-program` is set in `~/.gnupg/gpg-agent.conf`, then `gpgconf --kill gpg-agent`.

## Development

```bash
make build       # Build the binary
make check       # Run fmt, vet, lint, test
make test-unit   # Unit tests
make test-e2e    # E2E tests (mock pass-cli, no auth needed)
make lint        # Run linters
make coverage    # Coverage report
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and [test/README.md](test/README.md) for test details.

## Project Structure

```
pinentry-proton/
├── cmd/pinentry-proton/       # Application entry point
├── internal/
│   ├── config/                # Configuration loading and matching
│   ├── protocol/              # Pinentry Assuan protocol implementation
│   └── protonpass/            # ProtonPass CLI integration
├── test/                      # E2E tests, integration tests, fixtures
├── config.example.yaml        # Example configuration
└── Makefile                   # Build automation
```

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — Technical architecture and design
- [SECURITY.md](SECURITY.md) — Security policy and threat model
- [CONTRIBUTING.md](CONTRIBUTING.md) — Contribution guidelines

## License

MIT License — See [LICENSE](LICENSE) file for details.
