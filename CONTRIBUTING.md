# Contributing

## Before You Start

For anything beyond a small bug fix, open an issue first so we can discuss whether it fits the project's scope.

## Setup

```bash
git clone https://github.com/damoun/pinentry-proton.git
cd pinentry-proton
go mod download
make check   # fmt, vet, lint, tests
```

Requirements: Go 1.21+, golangci-lint. `pass-cli` only needed for real-ProtonPass tests.

## Making Changes

- Branch from `main`
- Run `make check` before pushing — the pre-push hook runs it anyway
- Keep PRs focused; one change per PR

## Security Rules

This tool handles passwords. Hard rules:

- Never log passwords, even in debug mode
- Keep passwords as `[]byte`, never cast to `string`
- Call `ZeroBytes()` on password buffers when done
- Never put passwords in command-line arguments or error messages

See [SECURITY.md](SECURITY.md) for the threat model.

## Tests

- New code needs tests; aim to keep coverage above 75%
- Use the mock pass-cli in `test/testutil/` — no real ProtonPass credentials in tests
- `make test-unit` for fast feedback, `make test-e2e` for full flows

## Commits

Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`

## Dependencies

Standard library first. Any new dependency needs a clear justification in the PR.

## Questions?

Open an issue. For security issues, use [GitHub Security Advisories](https://github.com/damoun/pinentry-proton/security/advisories/new) instead.
