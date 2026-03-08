# Security Policy

## Reporting a Vulnerability

Open a [GitHub Security Advisory](https://github.com/damoun/pinentry-proton/security/advisories/new) (private). Include:

- What the vulnerability is
- How to reproduce it
- Impact

I'll respond as soon as I can and coordinate a fix before any public disclosure.

## Threat Model

### Protected Against

- **Shoulder surfing** — passwords retrieved automatically, no typing required
- **Keyloggers** — no keyboard input of sensitive data
- **Process listing** — passwords never in command-line arguments
- **Memory persistence** — sensitive data zeroed after use
- **Signal interruption** — cleanup runs on SIGINT/SIGTERM

### Not Protected Against

- **Compromised ProtonPass account** — an attacker with account access can retrieve the same passwords
- **Root / kernel access** — root can read any process memory; kernel-level attackers can do more
- **Memory dumps during active use** — password is briefly in memory while being forwarded to the agent
- **Compromised pass-cli binary** — trust the ProtonPass CLI you install

## Implementation Notes

**No secret logging** — passwords never appear in logs, stderr, or error messages, even in debug mode.

**Memory handling** — passwords stay as `[]byte` (not `string`) so they can be explicitly zeroed with `ZeroBytes()` after use. `bytes.TrimSpace` is used instead of `strings.TrimSpace` to avoid creating unzeroable string copies.

**No persistence** — passwords are never written to disk or cached. Retrieved on-demand from ProtonPass each time.

**Timeout** — password retrieval times out after 60 seconds (configurable) to avoid hangs.

## Configuration File

The config file contains vault/item references, not passwords. Still, restrict its permissions:

```bash
chmod 600 ~/.config/pinentry-proton/config.yaml
```

Don't commit it to version control.

## Dependencies

- `gopkg.in/yaml.v3` — YAML config parsing
- Standard library for everything else

Dependencies are pinned in `go.sum`. Verify with `go mod verify`.

## Running Security Checks

```bash
make check   # fmt, vet, golangci-lint (includes gosec), unit tests
```

## Known Limitations

- Requires an active local ProtonPass session (pass-cli must be authenticated)
- Not suitable for headless/unattended environments without pre-authenticated pass-cli
- Not suitable for multi-user systems with shared home directories
