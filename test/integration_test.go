package test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testPIN         = "424242"
	testItemURI     = "pass://KmDQwh8YtmA3hKFn1x4KucB4ZXBG4_GXKLKp9oRP6uf_jn8wTTjzjnnP7A92KdQXmLp4kvgBAertdUZgggtZhQ==/MYhqRQ1mT5yo-l0TUh_Dzm38QvCsegOdKU2OWemXRheOOVAuv46qq7UBf6gWX3ZfiMDoOKnlfpSSPzAKRR_BRg=="
	binaryTimeout   = 5 * time.Second
	responseTimeout = 2 * time.Second
)

// TestBinaryExists verifies the pinentry-proton binary is built
func TestBinaryExists(t *testing.T) {
	binPath := getPinentryBinaryPath(t)
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("Binary not found at %s. Run 'make build' first: %v", binPath, err)
	}
}

// TestBasicProtocol tests the basic Assuan protocol handshake
func TestBasicProtocol(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommand(t, ctx)
	defer cmd.Process.Kill()

	// Read greeting
	greeting := readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(greeting, "OK Proton Pass pinentry") {
		t.Errorf("Unexpected greeting: %s", greeting)
	}

	// Test GETINFO commands
	testCommand(t, stdin, stdout, "GETINFO version", "D ")
	testCommand(t, stdin, stdout, "GETINFO pid", "D ")
	testCommand(t, stdin, stdout, "GETINFO flavor", "D proton")

	// Clean shutdown
	writeCommand(t, stdin, "BYE")
	response := readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, "OK") {
		t.Errorf("Expected OK on BYE, got: %s", response)
	}

	cleanExit(t, cmd)
}

// TestSetAndGetOptions tests setting various pinentry options
func TestSetAndGetOptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommand(t, ctx)
	defer cmd.Process.Kill()

	readLine(t, stdout, responseTimeout) // greeting

	tests := []struct {
		command string
		expect  string
	}{
		{"SETDESC Testing SSH key passphrase", "OK"},
		{"SETPROMPT Passphrase:", "OK"},
		{"SETTITLE SSH Key Required", "OK"},
		{"SETKEYINFO ssh:12345678", "OK"},
		{"SETOK Continue", "OK"},
		{"SETCANCEL Abort", "OK"},
		{"SETNOTOK Deny", "OK"},
		{"SETERROR Invalid passphrase", "OK"},
	}

	for _, tt := range tests {
		testCommand(t, stdin, stdout, tt.command, tt.expect)
	}

	writeCommand(t, stdin, "BYE")
	readLine(t, stdout, responseTimeout)
	cleanExit(t, cmd)
}

// TestGetPinWithMockProtonPass tests GETPIN using a mock ProtonPass CLI
func TestGetPinWithMockProtonPass(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock pass binary
	mockPassDir := t.TempDir()
	mockPassPath := filepath.Join(mockPassDir, "pass")
	createMockPassCLI(t, mockPassPath, testPIN)

	// Create test config
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	createTestConfig(t, configPath, testItemURI)

	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommandWithEnv(t, ctx, map[string]string{
		"PINENTRY_PROTON_CONFIG": configPath,
		"PATH":                   mockPassDir + ":" + os.Getenv("PATH"),
	})
	defer cmd.Process.Kill()

	readLine(t, stdout, responseTimeout) // greeting

	// Set context for SSH
	testCommand(t, stdin, stdout, "SETDESC SSH key passphrase", "OK")
	testCommand(t, stdin, stdout, "SETPROMPT Passphrase:", "OK")

	// Request PIN
	writeCommand(t, stdin, "GETPIN")

	// Should get D-line with PIN
	response := readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, "D ") {
		t.Fatalf("Expected D response with PIN, got: %s", response)
	}

	// Extract PIN (it will be percent-encoded)
	pin := strings.TrimPrefix(response, "D ")
	if !strings.Contains(pin, testPIN) && pin != testPIN {
		// Check if it's percent-encoded
		decodedPIN := percentDecode(pin)
		if decodedPIN != testPIN {
			t.Errorf("Expected PIN %s, got: %s (decoded: %s)", testPIN, pin, decodedPIN)
		}
	}

	// Should get OK
	response = readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, "OK") {
		t.Errorf("Expected OK after PIN, got: %s", response)
	}

	writeCommand(t, stdin, "BYE")
	readLine(t, stdout, responseTimeout)
	cleanExit(t, cmd)
}

// TestGetPinWithGPGContext tests GETPIN with GPG-specific context
func TestGetPinWithGPGContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockPassDir := t.TempDir()
	mockPassPath := filepath.Join(mockPassDir, "pass")
	createMockPassCLI(t, mockPassPath, testPIN)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	createTestConfig(t, configPath, testItemURI)

	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommandWithEnv(t, ctx, map[string]string{
		"PINENTRY_PROTON_CONFIG": configPath,
		"PATH":                   mockPassDir + ":" + os.Getenv("PATH"),
	})
	defer cmd.Process.Kill()

	readLine(t, stdout, responseTimeout) // greeting

	// Set GPG context
	testCommand(t, stdin, stdout, "SETDESC Please enter passphrase for GPG key", "OK")
	testCommand(t, stdin, stdout, "SETKEYINFO gpg:ABCD1234EFGH5678", "OK")

	// Request PIN
	writeCommand(t, stdin, "GETPIN")

	// Read response
	response := readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, "D ") {
		t.Fatalf("Expected D response, got: %s", response)
	}

	// OK response
	response = readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, "OK") {
		t.Errorf("Expected OK, got: %s", response)
	}

	writeCommand(t, stdin, "BYE")
	readLine(t, stdout, responseTimeout)
	cleanExit(t, cmd)
}

// TestCancelGetPin tests cancelling a GETPIN request
func TestCancelGetPin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockPassDir := t.TempDir()
	mockPassPath := filepath.Join(mockPassDir, "pass")
	// Mock that simulates user cancellation (exits with error)
	createMockPassCLIWithError(t, mockPassPath)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	createTestConfig(t, configPath, testItemURI)

	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommandWithEnv(t, ctx, map[string]string{
		"PINENTRY_PROTON_CONFIG": configPath,
		"PATH":                   mockPassDir + ":" + os.Getenv("PATH"),
	})
	defer cmd.Process.Kill()

	readLine(t, stdout, responseTimeout) // greeting

	testCommand(t, stdin, stdout, "SETDESC SSH key", "OK")

	// Request PIN (will fail due to mock error)
	writeCommand(t, stdin, "GETPIN")

	// Should get ERR response
	response := readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, "ERR") {
		t.Errorf("Expected ERR response when ProtonPass fails, got: %s", response)
	}

	writeCommand(t, stdin, "BYE")
	readLine(t, stdout, responseTimeout)
	cleanExit(t, cmd)
}

// TestInvalidCommands tests error handling for invalid commands
func TestInvalidCommands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommand(t, ctx)
	defer cmd.Process.Kill()

	readLine(t, stdout, responseTimeout) // greeting

	tests := []struct {
		command string
		expect  string
	}{
		{"INVALIDCOMMAND", "ERR"},
		{"GETINFO invalid_option", "ERR"},
		// OPTION currently accepts any option (returns OK)
		// This matches many simple pinentry implementations
	}

	for _, tt := range tests {
		writeCommand(t, stdin, tt.command)
		response := readLine(t, stdout, responseTimeout)
		if !strings.HasPrefix(response, tt.expect) {
			t.Errorf("Command %s: expected %s, got: %s", tt.command, tt.expect, response)
		}
	}

	writeCommand(t, stdin, "BYE")
	readLine(t, stdout, responseTimeout)
	cleanExit(t, cmd)
}

// TestMultipleGetPin tests multiple GETPIN requests in one session
func TestMultipleGetPin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockPassDir := t.TempDir()
	mockPassPath := filepath.Join(mockPassDir, "pass")
	createMockPassCLI(t, mockPassPath, testPIN)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	createTestConfig(t, configPath, testItemURI)

	ctx, cancel := context.WithTimeout(context.Background(), binaryTimeout)
	defer cancel()

	cmd, stdin, stdout := setupPinentryCommandWithEnv(t, ctx, map[string]string{
		"PINENTRY_PROTON_CONFIG": configPath,
		"PATH":                   mockPassDir + ":" + os.Getenv("PATH"),
	})
	defer cmd.Process.Kill()

	readLine(t, stdout, responseTimeout) // greeting

	// First GETPIN
	testCommand(t, stdin, stdout, "SETDESC First request", "OK")
	writeCommand(t, stdin, "GETPIN")
	readLine(t, stdout, responseTimeout) // D line
	testCommand(t, stdin, stdout, "", "OK")

	// Second GETPIN
	testCommand(t, stdin, stdout, "SETDESC Second request", "OK")
	writeCommand(t, stdin, "GETPIN")
	readLine(t, stdout, responseTimeout) // D line
	testCommand(t, stdin, stdout, "", "OK")

	writeCommand(t, stdin, "BYE")
	readLine(t, stdout, responseTimeout)
	cleanExit(t, cmd)
}

// Helper Functions

func getPinentryBinaryPath(t *testing.T) string {
	t.Helper()
	// Look for binary in project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Go up from test/ to project root
	projectRoot := filepath.Dir(wd)
	return filepath.Join(projectRoot, "pinentry-proton")
}

func cleanExit(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer waitCancel()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Logf("Command exited with error: %v", err)
		}
	case <-waitCtx.Done():
		t.Log("Command did not exit in time, killing")
		cmd.Process.Kill()
	}
}

func setupPinentryCommand(t *testing.T, ctx context.Context) (*exec.Cmd, io.Writer, io.Reader) {
	t.Helper()
	return setupPinentryCommandWithEnv(t, ctx, nil)
}

func setupPinentryCommandWithEnv(t *testing.T, ctx context.Context, env map[string]string) (*exec.Cmd, io.Writer, io.Reader) {
	t.Helper()

	binPath := getPinentryBinaryPath(t)
	cmd := exec.CommandContext(ctx, binPath)

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	// Capture stderr for debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start pinentry: %v", err)
	}

	t.Cleanup(func() {
		if stderr.Len() > 0 {
			t.Logf("Stderr output:\n%s", stderr.String())
		}
	})

	return cmd, stdin, stdout
}

func writeCommand(t *testing.T, w io.Writer, cmd string) {
	t.Helper()
	_, err := fmt.Fprintf(w, "%s\n", cmd)
	if err != nil {
		t.Fatalf("Failed to write command: %v", err)
	}
}

func readLine(t *testing.T, r io.Reader, timeout time.Duration) string {
	t.Helper()

	result := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(r)
		if scanner.Scan() {
			result <- scanner.Text()
		} else if err := scanner.Err(); err != nil {
			errChan <- err
		} else {
			errChan <- io.EOF
		}
	}()

	select {
	case line := <-result:
		return line
	case err := <-errChan:
		t.Fatalf("Failed to read line: %v", err)
		return ""
	case <-time.After(timeout):
		t.Fatal("Timeout reading line")
		return ""
	}
}

func testCommand(t *testing.T, stdin io.Writer, stdout io.Reader, cmd, expectedPrefix string) {
	t.Helper()
	if cmd != "" {
		writeCommand(t, stdin, cmd)
	}
	response := readLine(t, stdout, responseTimeout)
	if !strings.HasPrefix(response, expectedPrefix) {
		t.Errorf("Command %q: expected prefix %q, got: %s", cmd, expectedPrefix, response)
	}
}

func createTestConfig(t *testing.T, path, itemURI string) {
	t.Helper()
	config := fmt.Sprintf(`default_item: "%s"
timeout: 60
mappings:
  - name: "Test SSH"
    item: "%s"
    match:
      description: "ssh"
  - name: "Test GPG"
    item: "%s"
    match:
      description: "gpg"
`, itemURI, itemURI, itemURI)

	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
}

func createMockPassCLI(t *testing.T, path, password string) {
	t.Helper()
	script := fmt.Sprintf(`#!/bin/sh
# Mock ProtonPass CLI that returns test password
echo "%s"
exit 0
`, password)

	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
}

func createMockPassCLIWithError(t *testing.T, path string) {
	t.Helper()
	script := `#!/bin/sh
# Mock ProtonPass CLI that fails
echo "Error: user cancelled" >&2
exit 1
`
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
}

func percentDecode(s string) string {
	// Simple percent decoder for testing
	result := strings.ReplaceAll(s, "%20", " ")
	result = strings.ReplaceAll(result, "%0A", "\n")
	result = strings.ReplaceAll(result, "%0D", "\r")
	return result
}
