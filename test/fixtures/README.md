# Test Fixtures

This directory contains test keys and configuration for integration testing.

## Test Keys

All test keys use the PIN/passphrase: **424242**

### GPG Test Key
- Key ID: `ED25519 Test Key <test@pinentry-proton.local>`
- Purpose: Testing GPG agent integration
- Passphrase: 424242

### SSH Test Key
- Type: Ed25519
- Purpose: Testing SSH agent integration
- Passphrase: 424242

## ProtonPass Test Item

Default test item URI:
```
pass://KmDQwh8YtmA3hKFn1x4KucB4ZXBG4_GXKLKp9oRP6uf_jn8wTTjzjnnP7A92KdQXmLp4kvgBAertdUZgggtZhQ==/MYhqRQ1mT5yo-l0TUh_Dzm38QvCsegOdKU2OWemXRheOOVAuv46qq7UBf6gWX3ZfiMDoOKnlfpSSPzAKRR_BRg==
```

This item should return the password: `424242`

## Security Notice

**DO NOT USE THESE KEYS IN PRODUCTION**

These keys are for testing only and are intentionally published in this repository.
