#!/usr/bin/env bash
# Setup script for creating test GPG and SSH keys
# PIN/Passphrase: 424242

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="${SCRIPT_DIR}/fixtures"
GNUPGHOME="${FIXTURES_DIR}/gnupg"
SSH_DIR="${FIXTURES_DIR}/ssh"

echo "==> Creating test key infrastructure..."

# Clean up any existing test keys
rm -rf "${GNUPGHOME}" "${SSH_DIR}"
mkdir -p "${GNUPGHOME}" "${SSH_DIR}"
chmod 700 "${GNUPGHOME}" "${SSH_DIR}"

# Configure GPG to work in batch mode without agent
cat > "${GNUPGHOME}/gpg.conf" <<'GPGCONF'
# Test GPG configuration - no agent
no-tty
batch
pinentry-mode loopback
use-agent
GPGCONF

cat > "${GNUPGHOME}/gpg-agent.conf" <<'AGENTCONF'
# Minimal agent configuration for testing
allow-loopback-pinentry
default-cache-ttl 600
max-cache-ttl 7200
disable-scdaemon
log-file /tmp/gpg-agent-test.log
AGENTCONF

chmod 600 "${GNUPGHOME}/gpg.conf" "${GNUPGHOME}/gpg-agent.conf"

# Set GPG agent socket paths explicitly
export GPG_AGENT_INFO="${GNUPGHOME}/S.gpg-agent::1"

# GPG Test Key
echo ""
echo "==> Creating GPG test key (without passphrase initially)..."

# Create key generation script without passphrase
cat > "${GNUPGHOME}/gpg-key-script" <<'EOF'
%echo Generating test GPG key
Key-Type: EDDSA
Key-Curve: Ed25519
Key-Usage: sign,cert
Subkey-Type: ECDH
Subkey-Curve: Curve25519
Subkey-Usage: encrypt
Name-Real: Pinentry Proton Test
Name-Email: test@pinentry-proton.local
Expire-Date: 0
%no-protection
%commit
%echo Test GPG key created
EOF

# Generate key without passphrase (no agent needed)
gpg --homedir="${GNUPGHOME}" --batch --generate-key "${GNUPGHOME}/gpg-key-script"

echo "Adding passphrase to key..."
# Now add passphrase using gpg --passwd
# First, start the agent properly
gpgconf --homedir="${GNUPGHOME}" --launch gpg-agent 2>/dev/null || {
    # If gpgconf fails, try direct launch
    gpg-agent --homedir="${GNUPGHOME}" --daemon --allow-loopback-pinentry 2>/dev/null &
    sleep 2
}

# Get the key fingerprint
KEY_FP=$(gpg --homedir="${GNUPGHOME}" --list-keys --with-colons test@pinentry-proton.local 2>/dev/null | awk -F: '/^fpr:/ {print $10; exit}')

# Add passphrase using --passwd command
echo "424242" | gpg --homedir="${GNUPGHOME}" --batch --pinentry-mode loopback --passphrase-fd 0 --command-fd 0 --status-fd 2 --passwd "${KEY_FP}" 2>/dev/null || {
    echo "Note: Could not add passphrase automatically. Key created without passphrase."
    echo "You can add it manually with: gpg --homedir=${GNUPGHOME} --passwd ${KEY_FP}"
}

# Export the key for reference
gpg --homedir="${GNUPGHOME}" --armor --export test@pinentry-proton.local > "${FIXTURES_DIR}/test-gpg-public.asc"
echo "GPG public key exported to: ${FIXTURES_DIR}/test-gpg-public.asc"

# Get key ID
KEY_ID=$(gpg --homedir="${GNUPGHOME}" --list-keys --with-colons test@pinentry-proton.local | awk -F: '/^pub:/ {print $5}')
echo "GPG Key ID: ${KEY_ID}"

# SSH Test Key
echo ""
echo "==> Creating SSH test key..."
ssh-keygen -t ed25519 -f "${SSH_DIR}/test_ed25519" -N "424242" -C "test@pinentry-proton.local"
chmod 600 "${SSH_DIR}/test_ed25519"
chmod 644 "${SSH_DIR}/test_ed25519.pub"

echo ""
echo "==> Test keys created successfully!"
echo ""
echo "Test Key Information:"
echo "  GPG Key ID: ${KEY_ID}"
echo "  GPG Home: ${GNUPGHOME}"
echo "  SSH Key: ${SSH_DIR}/test_ed25519"
echo "  Passphrase for all keys: 424242"
echo ""
echo "To use these keys in tests:"
echo "  export GNUPGHOME=${GNUPGHOME}"
echo "  export SSH_KEY=${SSH_DIR}/test_ed25519"
