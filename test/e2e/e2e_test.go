package e2e

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

	"github.com/damoun/pinentry-proton/test/testutil"
)

const (
	testPassword = "424242"
	testVault    = "test-vault"
	testItem     = "test-item"
)

// TestE2E_FullProtocolFlow tests the complete pinentry protocol flow
func TestE2E_FullProtocolFlow(t *testing.T) {
	binary := getBinaryPath(t)
	mockCLI := testutil.CreateMockPassCLI(t, testVault, testItem, "password", testPassword)
	configPath, binDir := createE2EConfig(t, mockCLI, testVault, testItem)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
		fmt.Sprintf("PATH=%s:%s", binDir, os.Getenv("PATH")),
	)

	stdin, err := cmd.StdinPipe()
	testutil.AssertNoError(t, err, "Failed to create stdin pipe")

	stdout, err := cmd.StdoutPipe()
	testutil.AssertNoError(t, err, "Failed to create stdout pipe")

	err = cmd.Start()
	testutil.AssertNoError(t, err, "Failed to start pinentry")
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(stdout)

	// Read greeting
	if !scanner.Scan() {
		t.Fatal("Failed to read greeting")
	}
	greeting := scanner.Text()
	testutil.AssertProtocolOK(t, greeting)
	testutil.AssertContains(t, greeting, "Proton Pass pinentry", "Greeting should mention Proton Pass")

	// Send SET commands to establish context
	commands := []struct {
		cmd      string
		expected string
	}{
		{"SETDESC Test description", "OK"},
		{"SETPROMPT PIN:", "OK"},
		{"SETTITLE Passphrase", "OK"},
		{"SETKEYINFO test-key", "OK"},
	}

	for _, tc := range commands {
		fmt.Fprintf(stdin, "%s\n", tc.cmd)
		if !scanner.Scan() {
			t.Fatalf("Failed to read response for %s", tc.cmd)
		}
		response := scanner.Text()
		testutil.AssertProtocolOK(t, response)
	}

	// Send GETPIN
	fmt.Fprintf(stdin, "GETPIN\n")

	// Read data response
	if !scanner.Scan() {
		t.Fatal("Failed to read GETPIN data response")
	}
	dataResponse := scanner.Text()
	testutil.AssertProtocolData(t, dataResponse)

	// Decode the percent-encoded password
	encodedPassword := strings.TrimPrefix(dataResponse, "D ")
	decodedPassword := percentDecode(encodedPassword)
	testutil.AssertEqual(t, testPassword, decodedPassword, "Password should match")

	// Read OK response
	if !scanner.Scan() {
		t.Fatal("Failed to read GETPIN OK response")
	}
	testutil.AssertProtocolOK(t, scanner.Text())

	// Send BYE
	fmt.Fprintf(stdin, "BYE\n")
	if !scanner.Scan() {
		t.Fatal("Failed to read BYE response")
	}
	testutil.AssertProtocolOK(t, scanner.Text())

	stdin.Close()
	cmd.Wait()
}

// TestE2E_GPGWorkflow tests GPG-specific context and matching
func TestE2E_GPGWorkflow(t *testing.T) {
	binary := getBinaryPath(t)
	mockCLI := testutil.CreateMockPassCLI(t, "gpg", "signing-key", "password", testPassword)

	// Create config with GPG mapping
	config := testutil.TestConfig{
		DefaultItem: fmt.Sprintf("pass://%s/%s/password", testVault, testItem),
		Timeout:     60,
		Mappings: []testutil.TestMapping{
			{
				Name: "GPG Signing Key",
				Item: "pass://gpg/signing-key/password",
				Match: testutil.TestMatchCriteria{
					Description: "gpg: signing key",
				},
			},
		},
	}

	tmpDir, _ := testutil.SetupTestEnvironment(t)
	configPath := testutil.CreateTestConfig(t, config)

	// Set PATH to include mock CLI
	mockCLIDir := filepath.Dir(mockCLI)
	newPath := fmt.Sprintf("%s:%s", mockCLIDir, os.Getenv("PATH"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
		fmt.Sprintf("PATH=%s", newPath),
	)
	cmd.Dir = tmpDir

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Scan() // greeting

	// Send GPG-like context
	fmt.Fprintf(stdin, "SETDESC gpg: signing key ABCD1234\n")
	scanner.Scan()
	fmt.Fprintf(stdin, "SETTITLE Passphrase\n")
	scanner.Scan()

	// GETPIN should match the GPG mapping
	fmt.Fprintf(stdin, "GETPIN\n")
	scanner.Scan() // data
	dataLine := scanner.Text()
	scanner.Scan() // OK

	decodedPassword := percentDecode(strings.TrimPrefix(dataLine, "D "))
	testutil.AssertEqual(t, testPassword, decodedPassword, "Should retrieve GPG key password")

	fmt.Fprintf(stdin, "BYE\n")
	scanner.Scan()
	stdin.Close()
	cmd.Wait()
}

// TestE2E_SSHWorkflow tests SSH-specific context and matching
func TestE2E_SSHWorkflow(t *testing.T) {
	binary := getBinaryPath(t)
	mockCLI := testutil.CreateMockPassCLI(t, "ssh", "github-key", "password", testPassword)

	config := testutil.TestConfig{
		DefaultItem: fmt.Sprintf("pass://%s/%s/password", testVault, testItem),
		Timeout:     60,
		Mappings: []testutil.TestMapping{
			{
				Name: "GitHub SSH Key",
				Item: "pass://ssh/github-key/password",
				Match: testutil.TestMatchCriteria{
					Description: "github",
				},
			},
		},
	}

	tmpDir, _ := testutil.SetupTestEnvironment(t)
	configPath := testutil.CreateTestConfig(t, config)
	mockCLIDir := filepath.Dir(mockCLI)
	newPath := fmt.Sprintf("%s:%s", mockCLIDir, os.Getenv("PATH"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
		fmt.Sprintf("PATH=%s", newPath),
	)
	cmd.Dir = tmpDir

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Scan() // greeting

	// Send SSH-like context
	fmt.Fprintf(stdin, "SETDESC Enter passphrase for ~/.ssh/id_ed25519 (github.com)\n")
	scanner.Scan()
	fmt.Fprintf(stdin, "SETPROMPT Passphrase:\n")
	scanner.Scan()

	fmt.Fprintf(stdin, "GETPIN\n")
	scanner.Scan() // data
	dataLine := scanner.Text()
	scanner.Scan() // OK

	decodedPassword := percentDecode(strings.TrimPrefix(dataLine, "D "))
	testutil.AssertEqual(t, testPassword, decodedPassword, "Should retrieve SSH key password")

	fmt.Fprintf(stdin, "BYE\n")
	scanner.Scan()
	stdin.Close()
	cmd.Wait()
}

// TestE2E_ContextMatching tests configuration matching logic
func TestE2E_ContextMatching(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expectedURI string
		vault       string
		item        string
	}{
		{
			name:        "first match wins",
			description: "gpg signing",
			expectedURI: "pass://gpg/key1/password",
			vault:       "gpg",
			item:        "key1",
		},
		{
			name:        "case insensitive",
			description: "GPG SIGNING",
			expectedURI: "pass://gpg/key1/password",
			vault:       "gpg",
			item:        "key1",
		},
		{
			name:        "fallback to default",
			description: "unknown context",
			expectedURI: "pass://default/item/password",
			vault:       "default",
			item:        "item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary := getBinaryPath(t)
			mockCLI := testutil.CreateMockPassCLI(t, tt.vault, tt.item, "password", testPassword)

			config := testutil.TestConfig{
				DefaultItem: "pass://default/item/password",
				Timeout:     60,
				Mappings: []testutil.TestMapping{
					{
						Name: "GPG Key 1",
						Item: "pass://gpg/key1/password",
						Match: testutil.TestMatchCriteria{
							Description: "gpg signing",
						},
					},
					{
						Name: "GPG Key 2",
						Item: "pass://gpg/key2/password",
						Match: testutil.TestMatchCriteria{
							Description: "gpg encryption",
						},
					},
				},
			}

			tmpDir, _ := testutil.SetupTestEnvironment(t)
			configPath := testutil.CreateTestConfig(t, config)
			mockCLIDir := filepath.Dir(mockCLI)
			newPath := fmt.Sprintf("%s:%s", mockCLIDir, os.Getenv("PATH"))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
				fmt.Sprintf("PATH=%s", newPath),
			)
			cmd.Dir = tmpDir

			stdin, _ := cmd.StdinPipe()
			stdout, _ := cmd.StdoutPipe()
			cmd.Start()
			defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

			scanner := bufio.NewScanner(stdout)
			scanner.Scan() // greeting

			fmt.Fprintf(stdin, "SETDESC %s\n", tt.description)
			scanner.Scan()

			fmt.Fprintf(stdin, "GETPIN\n")
			scanner.Scan() // data
			dataLine := scanner.Text()
			scanner.Scan() // OK

			decodedPassword := percentDecode(strings.TrimPrefix(dataLine, "D "))
			testutil.AssertEqual(t, testPassword, decodedPassword, "Password should match")

			fmt.Fprintf(stdin, "BYE\n")
			scanner.Scan()
			stdin.Close()
			cmd.Wait()
		})
	}
}

// TestE2E_ErrorRecovery tests error handling scenarios
func TestE2E_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) (mockCLI, configPath string)
	}{
		{
			name: "pass-cli error",
			setupFunc: func(t *testing.T) (string, string) {
				mockCLI := testutil.CreateMockPassCLIWithError(t, "Item not found")
				config := testutil.TestConfig{
					DefaultItem: fmt.Sprintf("pass://%s/%s/password", testVault, testItem),
					Timeout:     60,
				}
				configPath := testutil.CreateTestConfig(t, config)
				return mockCLI, configPath
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary := getBinaryPath(t)
			mockCLI, configPath := tt.setupFunc(t)

			mockCLIDir := filepath.Dir(mockCLI)
			newPath := fmt.Sprintf("%s:%s", mockCLIDir, os.Getenv("PATH"))

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
				fmt.Sprintf("PATH=%s", newPath),
			)

			stdin, _ := cmd.StdinPipe()
			stdout, _ := cmd.StdoutPipe()
			cmd.Start()
			defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

			scanner := bufio.NewScanner(stdout)
			scanner.Scan() // greeting

			fmt.Fprintf(stdin, "GETPIN\n")
			scanner.Scan()
			errorResponse := scanner.Text()

			// Should get an error response
			testutil.AssertProtocolError(t, errorResponse)

			fmt.Fprintf(stdin, "BYE\n")
			scanner.Scan()
			stdin.Close()
			cmd.Wait()
		})
	}
}

// TestE2E_MultipleRequestsSameSession tests multiple GETPIN in one session
func TestE2E_MultipleRequestsSameSession(t *testing.T) {
	binary := getBinaryPath(t)
	mockCLI := testutil.CreateMockPassCLI(t, testVault, testItem, "password", testPassword)
	configPath, binDir := createE2EConfig(t, mockCLI, testVault, testItem)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
		fmt.Sprintf("PATH=%s:%s", binDir, os.Getenv("PATH")),
	)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(stdout)
	scanner.Scan() // greeting

	// Request password 3 times
	for i := 0; i < 3; i++ {
		fmt.Fprintf(stdin, "GETPIN\n")
		scanner.Scan() // data
		dataLine := scanner.Text()
		testutil.AssertProtocolData(t, dataLine)

		decodedPassword := percentDecode(strings.TrimPrefix(dataLine, "D "))
		testutil.AssertEqual(t, testPassword, decodedPassword, fmt.Sprintf("Request %d password mismatch", i+1))

		scanner.Scan() // OK
		testutil.AssertProtocolOK(t, scanner.Text())
	}

	fmt.Fprintf(stdin, "BYE\n")
	scanner.Scan()
	stdin.Close()
	cmd.Wait()
}

// TestE2E_LongPassword tests handling of passwords >1KB
func TestE2E_LongPassword(t *testing.T) {
	binary := getBinaryPath(t)
	longPassword := strings.Repeat("a", 2048)
	mockCLI := testutil.CreateMockPassCLI(t, testVault, testItem, "password", longPassword)
	configPath, binDir := createE2EConfig(t, mockCLI, testVault, testItem)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
		fmt.Sprintf("PATH=%s:%s", binDir, os.Getenv("PATH")),
	)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(stdout)
	// Increase buffer size for long passwords
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	scanner.Scan() // greeting

	fmt.Fprintf(stdin, "GETPIN\n")
	scanner.Scan() // data
	dataLine := scanner.Text()

	decodedPassword := percentDecode(strings.TrimPrefix(dataLine, "D "))
	testutil.AssertEqual(t, len(longPassword), len(decodedPassword), "Password length should match")
	testutil.AssertEqual(t, longPassword, decodedPassword, "Long password should match")

	scanner.Scan() // OK
	fmt.Fprintf(stdin, "BYE\n")
	scanner.Scan()
	stdin.Close()
	cmd.Wait()
}

// TestE2E_SpecialCharacters tests passwords with special characters
func TestE2E_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"spaces", "pass word 123"},
		{"symbols", "p@ss!w#rd$%^&*()"},
		{"quotes", `pass"word'123`},
		// Note: newlines in passwords are not supported by the protocol
		// as the protocol is line-based
		{"unicode", "пароль日本語🔐"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary := getBinaryPath(t)
			mockCLI := testutil.CreateMockPassCLI(t, testVault, testItem, "password", tt.password)
			configPath, binDir := createE2EConfig(t, mockCLI, testVault, testItem)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binary)
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
				fmt.Sprintf("PATH=%s:%s", binDir, os.Getenv("PATH")),
			)

			stdin, _ := cmd.StdinPipe()
			stdout, _ := cmd.StdoutPipe()
			cmd.Start()
			defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

			scanner := bufio.NewScanner(stdout)
			scanner.Scan() // greeting

			fmt.Fprintf(stdin, "GETPIN\n")
			scanner.Scan() // data
			dataLine := scanner.Text()

			decodedPassword := percentDecode(strings.TrimPrefix(dataLine, "D "))
			testutil.AssertEqual(t, tt.password, decodedPassword, "Special character password should match")

			scanner.Scan() // OK
			fmt.Fprintf(stdin, "BYE\n")
			scanner.Scan()
			stdin.Close()
			cmd.Wait()
		})
	}
}

// Helper functions

func getBinaryPath(t *testing.T) string {
	t.Helper()

	// Look for binary in project root
	binary := filepath.Join("..", "..", "pinentry-proton")
	if _, err := os.Stat(binary); err != nil {
		t.Fatalf("Binary not found at %s. Run 'make build' first: %v", binary, err)
	}
	return binary
}

func createE2EConfig(t *testing.T, mockCLI, vault, item string) (string, string) {
	t.Helper()

	config := testutil.TestConfig{
		DefaultItem: fmt.Sprintf("pass://%s/%s/password", vault, item),
		Timeout:     60,
	}

	tmpDir, _ := testutil.SetupTestEnvironment(t)
	configPath := testutil.CreateTestConfig(t, config)

	// Create a bin directory with pass-cli symlink/wrapper
	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	passCLIPath := filepath.Join(binDir, "pass-cli")

	// Create a wrapper script
	wrapper := fmt.Sprintf("#!/bin/bash\nexec '%s' \"$@\"\n", mockCLI)
	os.WriteFile(passCLIPath, []byte(wrapper), 0755)

	return configPath, binDir
}

func percentDecode(s string) string {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			var b byte
			fmt.Sscanf(s[i+1:i+3], "%02X", &b)
			buf.WriteByte(b)
			i += 2
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String()
}

// readOutput reads all available output with a timeout
func readOutput(r io.Reader, timeout time.Duration) ([]byte, error) {
	done := make(chan struct{})
	var buf bytes.Buffer

	go func() {
		io.Copy(&buf, r)
		close(done)
	}()

	select {
	case <-done:
		return buf.Bytes(), nil
	case <-time.After(timeout):
		return buf.Bytes(), fmt.Errorf("timeout reading output")
	}
}
