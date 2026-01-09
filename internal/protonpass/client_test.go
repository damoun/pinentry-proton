package protonpass

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewClient verifies client initialization
func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.cliPath != "pass-cli" {
		t.Errorf("Expected cliPath to be 'pass-cli', got %q", client.cliPath)
	}
}

// TestSetCLIPath verifies the SetCLIPath method
func TestSetCLIPath(t *testing.T) {
	client := NewClient()
	customPath := "/custom/path/to/pass-cli"

	client.SetCLIPath(customPath)

	if client.cliPath != customPath {
		t.Errorf("Expected cliPath to be %q, got %q", customPath, client.cliPath)
	}
}

// TestRetrievePassword_ValidURI tests the happy path with valid URI
func TestRetrievePassword_ValidURI(t *testing.T) {
	testPassword := "test-password-123"
	mockCLI := createMockCLI(t, "test", "item", "password", testPassword)

	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()
	password, err := client.RetrievePassword(ctx, "pass://test/item/password")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(password) != testPassword {
		t.Errorf("Expected password %q, got %q", testPassword, string(password))
	}

	// Verify password was returned as bytes
	if len(password) != len(testPassword) {
		t.Errorf("Expected password length %d, got %d", len(testPassword), len(password))
	}
}

// TestRetrievePassword_DefaultField tests that 'password' is the default field
func TestRetrievePassword_DefaultField(t *testing.T) {
	testPassword := "default-field-password"
	mockCLI := createMockCLI(t, "vault", "item", "password", testPassword)

	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()

	// URI without field should default to 'password'
	password, err := client.RetrievePassword(ctx, "pass://vault/item")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(password) != testPassword {
		t.Errorf("Expected password %q, got %q", testPassword, string(password))
	}
}

// TestRetrievePassword_CustomField tests custom field extraction
func TestRetrievePassword_CustomField(t *testing.T) {
	tests := []struct {
		name     string
		vault    string
		item     string
		field    string
		uri      string
		password string
	}{
		{
			name:     "passphrase field",
			vault:    "gpg",
			item:     "key",
			field:    "passphrase",
			uri:      "pass://gpg/key/passphrase",
			password: "my-passphrase",
		},
		{
			name:     "pin field",
			vault:    "cards",
			item:     "credit-card",
			field:    "pin",
			uri:      "pass://cards/credit-card/pin",
			password: "1234",
		},
		{
			name:     "token field",
			vault:    "api",
			item:     "github",
			field:    "token",
			uri:      "pass://api/github/token",
			password: "ghp_token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := createMockCLI(t, tt.vault, tt.item, tt.field, tt.password)

			client := NewClient()
			client.SetCLIPath(mockCLI)

			ctx := context.Background()
			password, err := client.RetrievePassword(ctx, tt.uri)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if string(password) != tt.password {
				t.Errorf("Expected password %q, got %q", tt.password, string(password))
			}
		})
	}
}

// TestRetrievePassword_InvalidURIs tests various malformed URIs
func TestRetrievePassword_InvalidURIs(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{
			name: "missing pass:// prefix",
			uri:  "vault/item/password",
		},
		{
			name: "only vault",
			uri:  "pass://vault",
		},
		{
			name: "empty components",
			uri:  "pass:///",
		},
		{
			name: "double slashes",
			uri:  "pass://vault//item",
		},
		{
			name: "empty string",
			uri:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient()
			ctx := context.Background()

			_, err := client.RetrievePassword(ctx, tt.uri)

			if err == nil {
				t.Errorf("Expected error for URI %q, got nil", tt.uri)
			}

			if !strings.Contains(err.Error(), "invalid item URI format") {
				t.Errorf("Expected 'invalid item URI format' error, got: %v", err)
			}
		})
	}
}

// TestRetrievePassword_SpecialCharacters tests URIs with special characters
func TestRetrievePassword_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		vault    string
		item     string
		field    string
		password string
	}{
		{
			name:     "spaces in item name",
			vault:    "work",
			item:     "GitHub SSH Key",
			field:    "password",
			password: "secret",
		},
		{
			name:     "dashes in names",
			vault:    "my-vault",
			item:     "my-item",
			field:    "password",
			password: "pwd",
		},
		{
			name:     "underscores",
			vault:    "work_vault",
			item:     "ssh_key",
			field:    "password",
			password: "s3cr3t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Spaces need to be passed correctly to the shell script
			mockCLI := createMockCLI(t, tt.vault, tt.item, tt.field, tt.password)

			client := NewClient()
			client.SetCLIPath(mockCLI)

			ctx := context.Background()
			uri := fmt.Sprintf("pass://%s/%s/%s", tt.vault, tt.item, tt.field)
			password, err := client.RetrievePassword(ctx, uri)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if string(password) != tt.password {
				t.Errorf("Expected password %q, got %q", tt.password, string(password))
			}
		})
	}
}

// TestRetrievePassword_EmptyPassword tests empty password handling
func TestRetrievePassword_EmptyPassword(t *testing.T) {
	mockCLI := createMockCLI(t, "vault", "item", "password", "")

	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()
	_, err := client.RetrievePassword(ctx, "pass://vault/item/password")

	if err == nil {
		t.Fatal("Expected error for empty password, got nil")
	}

	if !strings.Contains(err.Error(), "empty password") {
		t.Errorf("Expected 'empty password' error, got: %v", err)
	}
}

// TestRetrievePassword_WithWhitespace tests whitespace trimming
func TestRetrievePassword_WithWhitespace(t *testing.T) {
	tests := []struct {
		name             string
		rawPassword      string
		expectedPassword string
	}{
		{
			name:             "leading whitespace",
			rawPassword:      "  password",
			expectedPassword: "password",
		},
		{
			name:             "trailing whitespace",
			rawPassword:      "password  ",
			expectedPassword: "password",
		},
		{
			name:             "leading and trailing",
			rawPassword:      "  password  ",
			expectedPassword: "password",
		},
		{
			name:             "newline",
			rawPassword:      "password\n",
			expectedPassword: "password",
		},
		{
			name:             "tab",
			rawPassword:      "password\t",
			expectedPassword: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCLI := createMockCLI(t, "vault", "item", "password", tt.rawPassword)

			client := NewClient()
			client.SetCLIPath(mockCLI)

			ctx := context.Background()
			password, err := client.RetrievePassword(ctx, "pass://vault/item/password")

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if string(password) != tt.expectedPassword {
				t.Errorf("Expected password %q, got %q", tt.expectedPassword, string(password))
			}
		})
	}
}

// TestRetrievePassword_LongPassword tests passwords >1KB
func TestRetrievePassword_LongPassword(t *testing.T) {
	// Create a password longer than 1KB
	longPassword := strings.Repeat("a", 2048)

	mockCLI := createMockCLI(t, "vault", "item", "password", longPassword)

	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()
	password, err := client.RetrievePassword(ctx, "pass://vault/item/password")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(password) != longPassword {
		t.Errorf("Password length mismatch: expected %d, got %d", len(longPassword), len(password))
	}
}

// TestRetrievePassword_BinaryData tests passwords with binary-like data
func TestRetrievePassword_BinaryData(t *testing.T) {
	// Password with special characters that might be problematic
	binaryPassword := "p@ssw0rd!#$%^&*()[]{}|\\:;\"'<>,.?/~`"

	mockCLI := createMockCLI(t, "vault", "item", "password", binaryPassword)

	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()
	password, err := client.RetrievePassword(ctx, "pass://vault/item/password")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if string(password) != binaryPassword {
		t.Errorf("Expected password %q, got %q", binaryPassword, string(password))
	}
}

// TestRetrievePassword_CLINotFound tests pass-cli not found error
func TestRetrievePassword_CLINotFound(t *testing.T) {
	client := NewClient()
	client.SetCLIPath("/nonexistent/path/to/pass-cli")

	ctx := context.Background()
	_, err := client.RetrievePassword(ctx, "pass://vault/item/password")

	if err == nil {
		t.Fatal("Expected error when CLI not found, got nil")
	}

	// Should contain execution failed or similar
	if !strings.Contains(err.Error(), "pass-cli execution failed") &&
		!strings.Contains(err.Error(), "executable file not found") {
		t.Errorf("Expected 'execution failed' or 'not found' error, got: %v", err)
	}
}

// TestRetrievePassword_CLIError tests pass-cli exit codes
func TestRetrievePassword_CLIError(t *testing.T) {
	mockCLI := createMockCLIWithError(t, "Item not found")

	client := NewClient()
	client.SetCLIPath(mockCLI)

	ctx := context.Background()
	_, err := client.RetrievePassword(ctx, "pass://vault/item/password")

	if err == nil {
		t.Fatal("Expected error from CLI, got nil")
	}

	if !strings.Contains(err.Error(), "pass-cli error") {
		t.Errorf("Expected 'pass-cli error', got: %v", err)
	}
}

// TestRetrievePassword_ContextCancellation tests context cancellation
func TestRetrievePassword_ContextCancellation(t *testing.T) {
	// Create a mock that takes a long time
	mockCLI := createMockCLIWithDelay(t, 5*time.Second)

	client := NewClient()
	client.SetCLIPath(mockCLI)

	// Create context that cancels immediately
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.RetrievePassword(ctx, "pass://vault/item/password")

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// Should be context deadline exceeded or killed
	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "signal: killed") {
		t.Logf("Got error: %v", err)
	}
}

// TestZeroBytes verifies memory zeroing
func TestZeroBytes(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "simple password",
			data: []byte("password123"),
		},
		{
			name: "empty slice",
			data: []byte{},
		},
		{
			name: "single byte",
			data: []byte{0xFF},
		},
		{
			name: "large slice",
			data: make([]byte, 1024*1024), // 1MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill with non-zero data
			for i := range tt.data {
				tt.data[i] = 0xFF
			}

			// Zero the bytes
			ZeroBytes(tt.data)

			// Verify all bytes are zero
			for i, b := range tt.data {
				if b != 0 {
					t.Errorf("Byte at index %d is not zero: got %d", i, b)
				}
			}
		})
	}
}

// TestZeroBytes_Idempotent tests that zeroing can be called multiple times
func TestZeroBytes_Idempotent(t *testing.T) {
	data := []byte("test-data")

	// Zero multiple times
	ZeroBytes(data)
	ZeroBytes(data)
	ZeroBytes(data)

	// Should still be all zeros
	for i, b := range data {
		if b != 0 {
			t.Errorf("Byte at index %d is not zero after multiple calls: got %d", i, b)
		}
	}
}

// Helper functions

func createMockCLI(t *testing.T, vault, item, field, password string) string {
	t.Helper()

	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "mock-pass-cli")

	// Escape special characters in the password for shell script
	escapedPassword := strings.ReplaceAll(password, `"`, `\"`)
	escapedPassword = strings.ReplaceAll(escapedPassword, `$`, `\$`)
	escapedPassword = strings.ReplaceAll(escapedPassword, "`", "\\`")

	script := fmt.Sprintf(`#!/bin/bash
# Mock pass-cli for testing

if [ "$1" = "item" ] && [ "$2" = "get" ]; then
    ITEM_REF="$3"
    FIELD="password"

    # Parse --field flag if present
    if [ "$4" = "--field" ]; then
        FIELD="$5"
    fi

    # Expected values
    EXPECTED_REF="%s/%s"
    EXPECTED_FIELD="%s"

    if [ "$ITEM_REF" = "$EXPECTED_REF" ] && [ "$FIELD" = "$EXPECTED_FIELD" ]; then
        echo "%s"
        exit 0
    fi
fi

echo "Error: Item not found" >&2
exit 1
`, vault, item, field, escapedPassword)

	if err := os.WriteFile(cliPath, []byte(script), 0755); err != nil {
		t.Fatalf("Failed to create mock CLI: %v", err)
	}

	return cliPath
}

func createMockCLIWithError(t *testing.T, errorMsg string) string {
	t.Helper()

	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "mock-pass-cli-error")

	script := fmt.Sprintf(`#!/bin/bash
echo "%s" >&2
exit 1
`, errorMsg)

	if err := os.WriteFile(cliPath, []byte(script), 0755); err != nil {
		t.Fatalf("Failed to create error mock CLI: %v", err)
	}

	return cliPath
}

func createMockCLIWithDelay(t *testing.T, delay time.Duration) string {
	t.Helper()

	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "mock-pass-cli-delay")

	script := fmt.Sprintf(`#!/bin/bash
sleep %d
echo "delayed-password"
exit 0
`, int(delay.Seconds())+1)

	if err := os.WriteFile(cliPath, []byte(script), 0755); err != nil {
		t.Fatalf("Failed to create delay mock CLI: %v", err)
	}

	return cliPath
}
