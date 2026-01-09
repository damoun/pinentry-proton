package protocol

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/damoun/pinentry-proton/internal/config"
)

func TestUnescapeArg(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"hello%20world", "hello world"},
		{"test%0A", "test\n"},
		{"%25", "%"},
		{"no%20escape%20needed", "no escape needed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := UnescapeArg(tt.input)
			if got != tt.want {
				t.Errorf("UnescapeArg(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeArg(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"hello world", "hello world"},
		{"test\n", "test%0A"},
		{"%", "%25"},
		{"test\r\n", "test%0D%0A"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EscapeArg(tt.input)
			if got != tt.want {
				t.Errorf("EscapeArg(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPercentEncode(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte("hello"), "hello"},
		{[]byte("hello world"), "hello world"},
		{[]byte("test\n"), "test%0A"},
		{[]byte("%"), "%25"},
		{[]byte{0x00, 0x01}, "%00%01"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := PercentEncode(tt.input)
			if got != tt.want {
				t.Errorf("PercentEncode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSessionReset(t *testing.T) {
	cfg := &config.Config{}
	session := NewSession(strings.NewReader(""), &bytes.Buffer{}, cfg)

	session.description = "test desc"
	session.prompt = "test prompt"
	session.title = "test title"
	session.error = "test error"
	session.keyInfo = "test keyinfo"

	session.reset()

	if session.description != "" || session.prompt != "" ||
		session.title != "" || session.error != "" || session.keyInfo != "" {
		t.Error("Session not properly reset")
	}
}

func TestProtocolCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOK   bool
		wantData bool
	}{
		{
			name:   "SETDESC command",
			input:  "SETDESC Test description\n",
			wantOK: true,
		},
		{
			name:   "SETPROMPT command",
			input:  "SETPROMPT PIN:\n",
			wantOK: true,
		},
		{
			name:   "SETTITLE command",
			input:  "SETTITLE Test Title\n",
			wantOK: true,
		},
		{
			name:   "SETERROR command",
			input:  "SETERROR Test error\n",
			wantOK: true,
		},
		{
			name:   "SETKEYINFO command",
			input:  "SETKEYINFO ABC123\n",
			wantOK: true,
		},
		{
			name:     "GETINFO version",
			input:    "GETINFO version\n",
			wantOK:   true,
			wantData: true,
		},
		{
			name:     "GETINFO pid",
			input:    "GETINFO pid\n",
			wantOK:   true,
			wantData: true,
		},
		{
			name:   "BYE command",
			input:  "BYE\n",
			wantOK: true,
		},
		{
			name:   "RESET command",
			input:  "RESET\n",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			cfg := &config.Config{}

			session := NewSession(input, output, cfg)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// Run in goroutine to avoid blocking on GETPIN
			done := make(chan error, 1)
			go func() {
				done <- session.Run(ctx)
			}()

			select {
			case <-done:
				result := output.String()

				// First line should be the initial OK
				lines := strings.Split(strings.TrimSpace(result), "\n")
				if len(lines) < 1 {
					t.Fatal("No output received")
				}

				if !strings.HasPrefix(lines[0], "OK") {
					t.Errorf("First line should be OK, got: %s", lines[0])
				}

				if tt.wantOK && len(lines) < 2 {
					t.Error("Expected OK response, got no additional output")
				}

				if tt.wantData {
					hasData := false
					for _, line := range lines {
						if strings.HasPrefix(line, "D ") {
							hasData = true
							break
						}
					}
					if !hasData {
						t.Error("Expected data response (D), got none")
					}
				}

			case <-ctx.Done():
				t.Fatal("Test timed out")
			}
		})
	}
}

func TestMultipleCommands(t *testing.T) {
	input := strings.NewReader(`SETDESC Test description
SETPROMPT PIN:
SETTITLE Test
BYE
`)
	output := &bytes.Buffer{}
	cfg := &config.Config{}

	session := NewSession(input, output, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := session.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Unexpected error: %v", err)
	}

	result := output.String()
	okCount := strings.Count(result, "OK")

	// Should have: initial OK + 3 command OKs + BYE OK = 5
	if okCount != 5 {
		t.Errorf("Expected 5 OK responses, got %d. Output:\n%s", okCount, result)
	}
}
