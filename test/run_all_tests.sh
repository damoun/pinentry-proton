#!/usr/bin/env bash
# Run all integration tests

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Pinentry-Proton Integration Test Suite   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Check if binary exists
if [ ! -f "${PROJECT_ROOT}/pinentry-proton" ]; then
    echo -e "${YELLOW}Building pinentry-proton...${NC}"
    cd "${PROJECT_ROOT}"
    make build || {
        echo -e "${RED}Build failed${NC}"
        exit 1
    }
    echo -e "${GREEN}✓ Build successful${NC}"
    echo ""
fi

# Track results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    local test_script="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Running: ${test_name}${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    if bash "${test_script}"; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo ""
        echo -e "${GREEN}✓ ${test_name} PASSED${NC}"
    else
        FAILED_TESTS+=("${test_name}")
        echo ""
        echo -e "${RED}✗ ${test_name} FAILED${NC}"
    fi
    echo ""
}

# Run Go tests
run_test "Go Unit and Integration Tests" "${PROJECT_ROOT}/test/run_go_tests.sh"

# Run GPG tests
run_test "GPG Integration Test" "${SCRIPT_DIR}/test_gpg.sh"

# Run SSH tests
run_test "SSH Integration Test" "${SCRIPT_DIR}/test_ssh.sh"

# Summary
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Test Summary                     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Total tests: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${RED}Failed: 0${NC}"
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     All tests passed! 🎉 🎉 🎉            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}Failed: ${#FAILED_TESTS[@]}${NC}"
    echo ""
    echo -e "${RED}Failed tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗${NC} ${test}"
    done
    echo ""
    exit 1
fi
