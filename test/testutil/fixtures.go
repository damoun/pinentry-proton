// Package testutil provides shared testing utilities for pinentry-proton tests
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestConfig represents a test configuration structure
type TestConfig struct {
	DefaultItem string
	Timeout     int
	Mappings    []TestMapping
}

// TestMapping represents a test mapping structure
type TestMapping struct {
	Name  string
	Item  string
	Match TestMatchCriteria
}

// TestMatchCriteria represents test match criteria
type TestMatchCriteria struct {
	Description string
	Prompt      string
	Title       string
	KeyInfo     string
}

// CreateTestConfig generates a YAML configuration file for testing
func CreateTestConfig(t *testing.T, config TestConfig) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("default_item: \"%s\"\n", config.DefaultItem))
	sb.WriteString(fmt.Sprintf("timeout: %d\n", config.Timeout))

	if len(config.Mappings) > 0 {
		sb.WriteString("\nmappings:\n")
		for _, mapping := range config.Mappings {
			sb.WriteString(fmt.Sprintf("  - name: \"%s\"\n", mapping.Name))
			sb.WriteString(fmt.Sprintf("    item: \"%s\"\n", mapping.Item))
			sb.WriteString("    match:\n")

			if mapping.Match.Description != "" {
				sb.WriteString(fmt.Sprintf("      description: \"%s\"\n", mapping.Match.Description))
			}
			if mapping.Match.Prompt != "" {
				sb.WriteString(fmt.Sprintf("      prompt: \"%s\"\n", mapping.Match.Prompt))
			}
			if mapping.Match.Title != "" {
				sb.WriteString(fmt.Sprintf("      title: \"%s\"\n", mapping.Match.Title))
			}
			if mapping.Match.KeyInfo != "" {
				sb.WriteString(fmt.Sprintf("      keyinfo: \"%s\"\n", mapping.Match.KeyInfo))
			}
		}
	}

	if err := os.WriteFile(configPath, []byte(sb.String()), 0600); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	return configPath
}

// SetupTestEnvironment creates a temporary test environment with common setup
func SetupTestEnvironment(t *testing.T) (tempDir string, cleanup func()) {
	t.Helper()

	tempDir = t.TempDir()

	// Store original environment variables
	origEnv := map[string]string{
		"PINENTRY_PROTON_CONFIG": os.Getenv("PINENTRY_PROTON_CONFIG"),
		"PINENTRY_PROTON_DEBUG":  os.Getenv("PINENTRY_PROTON_DEBUG"),
		"PATH":                   os.Getenv("PATH"),
	}

	cleanup = func() {
		// Restore original environment
		for key, val := range origEnv {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}

	t.Cleanup(cleanup)

	return tempDir, cleanup
}

// AssertProtocolResponse checks if a protocol response matches expected format
func AssertProtocolResponse(t *testing.T, response, expectedPrefix string) {
	t.Helper()

	if !strings.HasPrefix(response, expectedPrefix) {
		t.Errorf("Expected response to start with %q, got %q", expectedPrefix, response)
	}
}

// AssertProtocolOK checks if a protocol response is OK
func AssertProtocolOK(t *testing.T, response string) {
	t.Helper()
	AssertProtocolResponse(t, response, "OK")
}

// AssertProtocolError checks if a protocol response is an error
func AssertProtocolError(t *testing.T, response string) {
	t.Helper()
	AssertProtocolResponse(t, response, "ERR")
}

// AssertProtocolData checks if a protocol response is data
func AssertProtocolData(t *testing.T, response string) {
	t.Helper()
	AssertProtocolResponse(t, response, "D ")
}

// WaitForCondition polls a condition function until it returns true or timeout
func WaitForCondition(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		<-ticker.C
	}

	return false
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertContains fails the test if haystack doesn't contain needle
func AssertContains(t *testing.T, haystack, needle string, message string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("%s: expected %q to contain %q", message, haystack, needle)
	}
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	tmpFile.Close()
	return tmpFile.Name()
}

// MakeExecutable makes a file executable
func MakeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatalf("Failed to make file executable: %v", err)
	}
}
