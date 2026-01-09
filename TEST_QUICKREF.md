# Quick Test Reference

## Test PIN and Item

```
PIN: 424242
Item: pass://KmDQwh8YtmA3hKFn1x4KucB4ZXBG4_GXKLKp9oRP6uf_jn8wTTjzjnnP7A92KdQXmLp4kvgBAertdUZgggtZhQ==/MYhqRQ1mT5yo-l0TUh_Dzm38QvCsegOdKU2OWemXRheOOVAuv46qq7UBf6gWX3ZfiMDoOKnlfpSSPzAKRR_BRg==
```

## Quick Commands

```bash
# First time setup
make test-setup              # Create GPG and SSH test keys

# Run tests
make test                    # Unit tests only
make test-all                # Everything (requires ProtonPass)
make test-gpg                # GPG integration
make test-ssh                # SSH integration
go test -short ./...         # All Go tests (skip ProtonPass)

# Debug
export PINENTRY_PROTON_DEBUG=1
make test-gpg

# Manual test
echo -e "GETINFO version\nBYE" | ./pinentry-proton
```

## Test Files

```
test/
├── integration_test.go      # 8 Go integration tests
├── test_gpg.sh              # GPG end-to-end test
├── test_ssh.sh              # SSH end-to-end test
├── setup_test_keys.sh       # Create test keys
├── run_all_tests.sh         # Master runner
├── run_go_tests.sh          # Go test runner
└── fixtures/
    ├── test-config.yaml     # Test configuration
    ├── gnupg/               # GPG keyring (after setup)
    └── ssh/                 # SSH keys (after setup)
```

## Documentation

- `TESTING.md` - Comprehensive guide with test suite summary
- `test/README.md` - Test directory guide
- `.github/workflows/ci.yml` - CI/CD workflow

## Requirements for Full Tests

1. ✅ Build binary: `make build`
2. ✅ Create test keys: `make test-setup`
3. ✅ Install ProtonPass CLI
4. ✅ Authenticate: `pass auth login`
5. ✅ Create test item with PIN 424242
6. ✅ Run: `make test-all`

## CI/CD (No ProtonPass)

```bash
make build
go test -short ./...
```
