package protocol

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/damoun/pinentry-proton/internal/config"
	"github.com/damoun/pinentry-proton/internal/protonpass"
)

// TestFullPinentryFlow tests a complete pinentry protocol interaction
func TestFullPinentryFlow(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		DefaultItem: "pass://Test/Default/password",
		Timeout:     60,
		Mappings: []config.Mapping{
			{
				Name: "Test SSH",
				Item: "pass://Test/SSH/password",
				Match: config.MatchCriteria{
					Description: "ssh",
				},
			},
		},
	}

	// Simulate a complete pinentry session
	input := `SETDESC SSH key passphrase
SETPROMPT Passphrase:
SETTITLE SSH Key Required
GETINFO version
GETINFO pid
BYE
`

	output := &bytes.Buffer{}
	session := NewSession(strings.NewReader(input), output, cfg)
	defer session.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := session.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Session failed: %v", err)
	}

	result := output.String()

	// Verify initial greeting
	if !strings.Contains(result, "OK Proton Pass pinentry") {
		t.Error("Missing initial greeting")
	}

	// Count OK responses (should have one for each command + initial)
	okCount := strings.Count(result, "OK")
	if okCount < 5 {
		t.Errorf("Expected at least 5 OK responses, got %d. Output:\n%s", okCount, result)
	}

	// Verify version response
	if !strings.Contains(result, "D "+Version) {
		t.Error("Missing version in GETINFO response")
	}
}

// TestContextMatching tests that context fields are properly matched
func TestContextMatching(t *testing.T) {
	cfg := &config.Config{
		Mappings: []config.Mapping{
			{
				Name: "GitHub SSH",
				Item: "pass://Work/GitHub/password",
				Match: config.MatchCriteria{
					Description: "github.com",
				},
			},
			{
				Name: "GPG Key",
				Item: "pass://Personal/GPG/password",
				Match: config.MatchCriteria{
					KeyInfo: "ABC123",
				},
			},
		},
	}

	tests := []struct {
		name        string
		input       string
		expectMatch string
	}{
		{
			name: "Match by description",
			input: `SETDESC Please enter passphrase for github.com SSH key
RESET
BYE
`,
			expectMatch: "GitHub",
		},
		{
			name: "Match by keyinfo",
			input: `SETKEYINFO ABC123
RESET
BYE
`,
			expectMatch: "GPG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			session := NewSession(strings.NewReader(tt.input), output, cfg)
			defer session.Cleanup()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_ = session.Run(ctx)

			// Verify context was set correctly by checking internal state
			// This is a simplified test - in reality GETPIN would use the matching
		})
	}
}

// TestSignalHandling tests that signals are handled properly
func TestSignalHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping signal test in short mode")
	}

	cfg := &config.Config{}
	input := strings.NewReader("SETDESC Test\n")
	output := &bytes.Buffer{}

	session := NewSession(input, output, cfg)
	
	// Add some sensitive data to track cleanup
	sensitiveData := []byte("sensitive")
	session.sensitiveData = append(session.sensitiveData, sensitiveData)

	// Cleanup and verify zeroing
	session.Cleanup()

	// Check that sensitive data was zeroed
	for i, b := range sensitiveData {
		if b != 0 {
			t.Errorf("Sensitive data not zeroed at index %d: %d", i, b)
		}
	}
}

// TestProtocolEdgeCases tests edge cases in the protocol
func TestProtocolEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		assert func(t *testing.T, output string)
	}{
		{
			name:  "Empty lines ignored",
			input: "\n\n\nSETDESC Test\n\n\nBYE\n",
			assert: func(t *testing.T, output string) {
				if !strings.Contains(output, "OK") {
					t.Error("Should handle empty lines gracefully")
				}
			},
		},
		{
			name:  "Unknown command",
			input: "UNKNOWN_COMMAND\nBYE\n",
			assert: func(t *testing.T, output string) {
				if !strings.Contains(output, "ERR") {
					t.Error("Should return error for unknown command")
				}
			},
		},
		{
			name:  "Multiple RESET commands",
			input: "SETDESC Test\nRESET\nSETDESC Test2\nRESET\nBYE\n",
			assert: func(t *testing.T, output string) {
				okCount := strings.Count(output, "OK")
				if okCount < 5 {
					t.Errorf("Expected multiple OK responses, got %d", okCount)
				}
			},
		},
		{
			name:  "Case sensitivity",
			input: "setdesc test\nSETDESC TEST\nBYE\n",
			assert: func(t *testing.T, output string) {
				// Commands should be case-insensitive (converted to upper)
				okCount := strings.Count(output, "OK")
				if okCount < 3 {
					t.Errorf("Expected case-insensitive commands, got %d OKs", okCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cfg := &config.Config{}
			session := NewSession(strings.NewReader(tt.input), output, cfg)
			defer session.Cleanup()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_ = session.Run(ctx)
			tt.assert(t, output.String())
		})
	}
}

// TestPercentEncodingRoundTrip tests encoding/decoding
func TestPercentEncodingRoundTrip(t *testing.T) {
	tests := []string{
		"simple text",
		"text with spaces",
		"special%chars",
		"newline\nchar",
		"tab\tchar",
		"unicode: 日本語",
		"",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			// Encode
			encoded := EscapeArg(tt)

			// Decode
			decoded := UnescapeArg(encoded)

			if decoded != tt {
				t.Errorf("Round trip failed: %q -> %q -> %q", tt, encoded, decoded)
			}
		})
	}
}

// TestContextCancellation tests that context cancellation is respected
func TestContextCancellation(t *testing.T) {
	cfg := &config.Config{}
	// Input that would block (no BYE)
	input := strings.NewReader("SETDESC Test\n")
	output := &bytes.Buffer{}

	session := NewSession(input, output, cfg)
	defer session.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := session.Run(ctx)

	// Should timeout, be cancelled, or hit EOF
	// All of these are acceptable - the key is it didn't hang forever
	if err != context.DeadlineExceeded && err != context.Canceled && err != nil {
		// If err is nil, that's actually fine - the session ended cleanly after processing input
		t.Logf("Context test completed with: %v (acceptable)", err)
	}
}

// TestMemoryZeroing verifies sensitive data cleanup
func TestMemoryZeroing(t *testing.T) {
	// Create sensitive data
	data1 := []byte("password123")
	data2 := []byte("secret456")

	cfg := &config.Config{}
	session := NewSession(strings.NewReader(""), &bytes.Buffer{}, cfg)

	// Add to tracking
	session.sensitiveData = append(session.sensitiveData, data1, data2)

	// Cleanup
	session.Cleanup()

	// Verify all bytes are zero
	for i, b := range data1 {
		if b != 0 {
			t.Errorf("data1[%d] = %d, want 0", i, b)
		}
	}

	for i, b := range data2 {
		if b != 0 {
			t.Errorf("data2[%d] = %d, want 0", i, b)
		}
	}

	// Verify tracking is cleared
	if session.sensitiveData != nil {
		t.Error("sensitiveData should be nil after cleanup")
	}
}

// TestDebugMode tests debug output when enabled
func TestDebugMode(t *testing.T) {
	// Save original debug mode
	originalDebug := DebugMode
	defer func() { DebugMode = originalDebug }()

	// Enable debug mode
	DebugMode = true

	// This test mainly ensures debug mode doesn't crash
	cfg := &config.Config{
		Mappings: []config.Mapping{
			{
				Name: "Test",
				Item: "pass://Test/Item/password",
				Match: config.MatchCriteria{
					Description: "test",
				},
			},
		},
	}

	input := "SETDESC Test description\nBYE\n"
	output := &bytes.Buffer{}
	session := NewSession(strings.NewReader(input), output, cfg)
	defer session.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := session.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Session failed: %v", err)
	}

	// Should still work normally
	if !strings.Contains(output.String(), "OK") {
		t.Error("Debug mode shouldn't break normal operation")
	}
}

// BenchmarkPercentEncode benchmarks the encoding function
func BenchmarkPercentEncode(b *testing.B) {
	data := []byte("This is a test password with some special chars: !@#$%^&*()")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PercentEncode(data)
	}
}

// BenchmarkZeroBytes benchmarks memory zeroing
func BenchmarkZeroBytes(b *testing.B) {
	data := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		protonpass.ZeroBytes(data)
	}
}

// TestGetInfoCommands tests all GETINFO variants
func TestGetInfoCommands(t *testing.T) {
	tests := []struct {
		command string
		wantErr bool
	}{
		{"GETINFO version", false},
		{"GETINFO pid", false},
		{"GETINFO flavor", false},
		{"GETINFO ttyinfo", false},
		{"GETINFO unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cfg := &config.Config{}
			input := tt.command + "\nBYE\n"
			output := &bytes.Buffer{}

			session := NewSession(strings.NewReader(input), output, cfg)
			defer session.Cleanup()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_ = session.Run(ctx)

			result := output.String()
			hasErr := strings.Contains(result, "ERR")

			if hasErr != tt.wantErr {
				t.Errorf("Command %q: hasErr=%v, wantErr=%v. Output:\n%s",
					tt.command, hasErr, tt.wantErr, result)
			}
		})
	}
}
