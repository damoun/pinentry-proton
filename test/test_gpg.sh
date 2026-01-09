#!/usr/bin/env bash
# End-to-end test with GPG using test key
# Requires: gpg, pinentry-proton binary, test keys setup

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="${SCRIPT_DIR}/fixtures"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"
BINARY="${PROJECT_ROOT}/pinentry-proton"

# Test configuration
export GNUPGHOME="${FIXTURES_DIR}/gnupg"
export PINENTRY_PROTON_CONFIG="${FIXTURES_DIR}/test-config.yaml"
TEST_PIN="424242"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}==> GPG Integration Test${NC}"
echo ""

# Check prerequisites
if [ ! -f "${BINARY}" ]; then
    echo -e "${RED}Error: Binary not found at ${BINARY}${NC}"
    echo "Run 'make build' first"
    exit 1
fi

if [ ! -d "${GNUPGHOME}" ]; then
    echo -e "${YELLOW}Test keys not found. Running setup...${NC}"
    "${SCRIPT_DIR}/setup_test_keys.sh"
fi

# Configure GPG to use our pinentry
echo "pinentry-program ${BINARY}" > "${GNUPGHOME}/gpg-agent.conf"

# Kill any running gpg-agent
gpgconf --homedir="${GNUPGHOME}" --kill gpg-agent 2>/dev/null || true
sleep 1

# Get test key ID
KEY_ID=$(gpg --homedir="${GNUPGHOME}" --list-secret-keys --with-colons | awk -F: '/^sec:/ {print $5; exit}')
if [ -z "${KEY_ID}" ]; then
    echo -e "${RED}Error: No GPG test key found${NC}"
    exit 1
fi

echo "Using GPG key: ${KEY_ID}"
echo ""

# Create test message
TEST_MESSAGE="This is a test message for GPG signing."
TEST_FILE="${FIXTURES_DIR}/test-message.txt"
echo "${TEST_MESSAGE}" > "${TEST_FILE}"

echo -e "${YELLOW}Test 1: Sign a message${NC}"
echo "This will invoke pinentry-proton to retrieve the passphrase..."

# Sign the message (this will trigger pinentry)
if gpg --homedir="${GNUPGHOME}" --local-user "${KEY_ID}" --armor --sign "${TEST_FILE}" 2>&1; then
    echo -e "${GREEN}✓ Message signed successfully${NC}"
    SIGNED_FILE="${TEST_FILE}.asc"
    
    # Verify the signature
    echo ""
    echo -e "${YELLOW}Test 2: Verify signature${NC}"
    if gpg --homedir="${GNUPGHOME}" --verify "${SIGNED_FILE}" 2>&1 | grep -q "Good signature"; then
        echo -e "${GREEN}✓ Signature verified successfully${NC}"
    else
        echo -e "${RED}✗ Signature verification failed${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ Signing failed${NC}"
    echo "Check that:"
    echo "  1. ProtonPass CLI is installed and authenticated"
    echo "  2. The test item exists and contains PIN: ${TEST_PIN}"
    echo "  3. pinentry-proton binary works: echo 'GETINFO version' | ${BINARY}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Test 3: Encrypt and decrypt${NC}"

# Encrypt a message
ENCRYPTED_FILE="${FIXTURES_DIR}/test-encrypted.gpg"
if gpg --homedir="${GNUPGHOME}" --recipient "${KEY_ID}" --encrypt --output "${ENCRYPTED_FILE}" "${TEST_FILE}" 2>&1; then
    echo -e "${GREEN}✓ Message encrypted${NC}"
    
    # Decrypt (will trigger pinentry again)
    DECRYPTED_FILE="${FIXTURES_DIR}/test-decrypted.txt"
    if gpg --homedir="${GNUPGHOME}" --decrypt --output "${DECRYPTED_FILE}" "${ENCRYPTED_FILE}" 2>&1; then
        echo -e "${GREEN}✓ Message decrypted${NC}"
        
        # Verify content
        if diff -q "${TEST_FILE}" "${DECRYPTED_FILE}" >/dev/null; then
            echo -e "${GREEN}✓ Decrypted content matches original${NC}"
        else
            echo -e "${RED}✗ Decrypted content differs${NC}"
            exit 1
        fi
    else
        echo -e "${RED}✗ Decryption failed${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ Encryption failed${NC}"
    exit 1
fi

# Cleanup
rm -f "${TEST_FILE}" "${TEST_FILE}.asc" "${ENCRYPTED_FILE}" "${DECRYPTED_FILE}"
gpgconf --homedir="${GNUPGHOME}" --kill gpg-agent 2>/dev/null || true

echo ""
echo -e "${GREEN}==> All GPG tests passed! 🎉${NC}"
echo ""
echo "The pinentry-proton successfully:"
echo "  • Retrieved passphrase from ProtonPass"
echo "  • Provided it to GPG"
echo "  • Enabled signing, verification, encryption, and decryption"
