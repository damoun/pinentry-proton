# Tests

All tests run without a real ProtonPass installation — they use a mock `pass-cli`.

## Running

```bash
make test        # Unit + integration tests
make test-e2e    # E2E tests (mock pass-cli)
make test-unit   # Unit tests only (internal packages)
```

## Structure

- `integration_test.go` — Protocol tests that run the built binary and talk to it via stdin/stdout
- `e2e/e2e_test.go` — Full workflow tests using a mock pass-cli
- `testutil/` — Shared helpers: mock pass-cli, config fixtures
- `fixtures/` — Test configs and mock data

## Debugging

```bash
export PINENTRY_PROTON_DEBUG=1
echo -e "GETINFO version\nBYE" | ./pinentry-proton
```

## Security

Test keys in `fixtures/` use passphrase `424242` and are committed to the repo. Never use them outside of testing.
