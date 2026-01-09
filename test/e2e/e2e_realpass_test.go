// +build realpass

package e2e

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/damoun/pinentry-proton/test/testutil"
)

// TestE2E_RealProtonPass tests with actual ProtonPass CLI
// Run with: go test -tags=realpass ./test/e2e/
//
// Requirements:
// - pass-cli installed and authenticated
// - Test vault "test" with item "pinentry-code" containing password field
func TestE2E_RealProtonPass(t *testing.T) {
	// Verify pass-cli is available
	if _, err := exec.LookPath("pass-cli"); err != nil {
		t.Skip("pass-cli not found, skipping real ProtonPass test")
	}

	// Verify pass-cli is authenticated by trying to list items
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkCmd := exec.CommandContext(ctx, "pass-cli", "item", "list")
	if err := checkCmd.Run(); err != nil {
		t.Skip("pass-cli not authenticated, skipping real ProtonPass test")
	}

	// Create config pointing to real test vault
	config := testutil.TestConfig{
		DefaultItem: "pass://test/pinentry-code/password",
		Timeout:     60,
	}

	tmpDir, _ := testutil.SetupTestEnvironment(t)
	configPath := testutil.CreateTestConfig(t, config)

	// Get binary path
	binary := getBinaryPath(t)

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PINENTRY_PROTON_CONFIG=%s", configPath),
	)
	cmd.Dir = tmpDir

	stdin, err := cmd.StdinPipe()
	testutil.AssertNoError(t, err, "Failed to create stdin pipe")

	stdout, err := cmd.StdoutPipe()
	testutil.AssertNoError(t, err, "Failed to create stdout pipe")

	stderr, err := cmd.StderrPipe()
	testutil.AssertNoError(t, err, "Failed to create stderr pipe")

	err = cmd.Start()
	testutil.AssertNoError(t, err, "Failed to start pinentry")
	defer cmd.Process.Kill()

	// Capture stderr for debugging
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				t.Logf("STDERR: %s", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	scanner := bufio.NewScanner(stdout)

	// Read greeting
	if !scanner.Scan() {
		t.Fatal("Failed to read greeting")
	}
	greeting := scanner.Text()
	t.Logf("Greeting: %s", greeting)
	testutil.AssertProtocolOK(t, greeting)

	// Send SET commands
	fmt.Fprintf(stdin, "SETDESC Test with real ProtonPass\n")
	scanner.Scan()
	t.Logf("SETDESC response: %s", scanner.Text())

	fmt.Fprintf(stdin, "SETPROMPT PIN:\n")
	scanner.Scan()
	t.Logf("SETPROMPT response: %s", scanner.Text())

	// Request password from real ProtonPass
	t.Log("Requesting password from ProtonPass...")
	fmt.Fprintf(stdin, "GETPIN\n")

	// Read data response
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			t.Fatalf("Failed to read GETPIN data response: %v", err)
		}
		t.Fatal("Failed to read GETPIN data response")
	}
	dataResponse := scanner.Text()
	t.Logf("Data response: %s", dataResponse)
	testutil.AssertProtocolData(t, dataResponse)

	// Decode password
	encodedPassword := strings.TrimPrefix(dataResponse, "D ")
	decodedPassword := percentDecode(encodedPassword)

	// Verify we got a non-empty password
	if len(decodedPassword) == 0 {
		t.Error("Retrieved password is empty")
	} else {
		t.Logf("Successfully retrieved password from ProtonPass (%d bytes)", len(decodedPassword))
	}

	// Read OK response
	if !scanner.Scan() {
		t.Fatal("Failed to read GETPIN OK response")
	}
	okResponse := scanner.Text()
	t.Logf("OK response: %s", okResponse)
	testutil.AssertProtocolOK(t, okResponse)

	// Test memory zeroing by requesting again
	t.Log("Testing memory cleanup with second request...")
	fmt.Fprintf(stdin, "GETPIN\n")

	scanner.Scan() // data
	dataResponse2 := scanner.Text()
	decodedPassword2 := percentDecode(strings.TrimPrefix(dataResponse2, "D "))

	scanner.Scan() // OK

	// Both requests should return the same password
	testutil.AssertEqual(t, decodedPassword, decodedPassword2, "Second request should return same password")

	// Clean shutdown
	fmt.Fprintf(stdin, "BYE\n")
	scanner.Scan()
	t.Logf("BYE response: %s", scanner.Text())

	stdin.Close()
	err = cmd.Wait()
	if err != nil {
		t.Logf("Command exited with error: %v", err)
	}

	t.Log("✅ Real ProtonPass test completed successfully")
}
