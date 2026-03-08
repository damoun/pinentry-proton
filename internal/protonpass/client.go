// Package protonpass handles integration with ProtonPass CLI (pass-cli).
package protonpass

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	// DebugMode controls whether debug logging is enabled
	DebugMode = os.Getenv("PINENTRY_PROTON_DEBUG") == "1"
)

// Client handles ProtonPass CLI interactions
type Client struct {
	cliPath string
}

// NewClient creates a new ProtonPass client
func NewClient() *Client {
	return &Client{
		cliPath: "pass-cli",
	}
}

// SetCLIPath sets a custom path for the pass-cli executable
// This is primarily used for testing with mock implementations
func (c *Client) SetCLIPath(path string) {
	c.cliPath = path
}

// RetrievePassword retrieves a password from ProtonPass using pass-cli
func (c *Client) RetrievePassword(ctx context.Context, itemURI string) ([]byte, error) {
	// Parse the ProtonPass URI: pass://VAULT/ITEM/FIELD
	if DebugMode {
		log.Printf("[DEBUG] Retrieving password from: %s", itemURI)
	}

	// Validate URI format
	if !strings.HasPrefix(itemURI, "pass://") {
		return nil, fmt.Errorf("invalid item URI format: %s (expected: pass://vault/item[/field])", itemURI)
	}

	// Construct pass-cli command
	itemPath := strings.TrimPrefix(itemURI, "pass://")
	parts := strings.Split(itemPath, "/")

	// Validate we have at least vault and item
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid item URI format: %s (expected: pass://vault/item[/field])", itemURI)
	}

	// Validate vault and item are not empty
	if parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid item URI format: %s (vault and item cannot be empty)", itemURI)
	}
	// Determine if we're getting a specific field or default to password
	field := "password"
	if len(parts) >= 3 {
		if parts[2] == "" {
			return nil, fmt.Errorf("invalid item URI format: %s (field cannot be empty)", itemURI)
		}
		field = parts[2]
	}

	// Build the full URI with field embedded: pass://vault/item/field
	fullURI := "pass://" + strings.Join(parts[:2], "/") + "/" + field

	if DebugMode {
		log.Printf("[DEBUG] Full URI: %s", fullURI)
	}

	// Execute pass-cli to get the item, passing the full URI (field embedded in path)
	cmd := exec.CommandContext(ctx, c.cliPath, "item", "view", fullURI) //nolint:gosec // G204: fullURI from user config, cliPath controlled by app

	// Capture stdout only; stderr is captured separately to avoid corrupting the password
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	output, err := cmd.Output()
	if err != nil {
		stderrOutput := stderrBuf.String()
		if DebugMode {
			log.Printf("[DEBUG] pass-cli error: %v, stderr: %s", err, stderrOutput)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("pass-cli error (exit %d): %s", exitErr.ExitCode(), stderrOutput)
		}
		return nil, fmt.Errorf("pass-cli execution failed: %w", err)
	}

	// Trim whitespace
	password := []byte(strings.TrimSpace(string(output)))

	if len(password) == 0 {
		return nil, fmt.Errorf("empty password returned from ProtonPass item: %s", itemURI)
	}

	if DebugMode {
		log.Printf("[DEBUG] Successfully retrieved password (%d bytes)", len(password))
	}

	return password, nil
}

// ZeroBytes securely zeros a byte slice
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
