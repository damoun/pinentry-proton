// Package protocol implements the pinentry Assuan protocol.
package protocol

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/damoun/pinentry-proton/internal/config"
	"github.com/damoun/pinentry-proton/internal/protonpass"
)

const (
	// Version of the pinentry implementation
	Version = "1.0.0"
	// DefaultTimeout for GETPIN operations
	DefaultTimeout = 60 * time.Second
)

var (
	// DebugMode controls whether debug logging is enabled
	DebugMode = os.Getenv("PINENTRY_PROTON_DEBUG") == "1"
)

// Session manages the pinentry protocol session
type Session struct {
	reader  *bufio.Reader
	writer  io.Writer
	config  *config.Config
	client  *protonpass.Client
	timeout time.Duration

	// Current request context
	description string
	prompt      string
	title       string
	error       string
	keyInfo     string

	// Cleanup tracking
	sensitiveData [][]byte
}

// NewSession creates a new pinentry session
func NewSession(reader io.Reader, writer io.Writer, cfg *config.Config) *Session {
	return &Session{
		reader:        bufio.NewReader(reader),
		writer:        writer,
		config:        cfg,
		client:        protonpass.NewClient(),
		timeout:       DefaultTimeout,
		sensitiveData: make([][]byte, 0),
	}
}

// Run starts the pinentry protocol session
func (s *Session) Run(ctx context.Context) error {
	// Send initial OK with version
	if err := s.writeOK(fmt.Sprintf("Proton Pass pinentry v%s ready", Version)); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse command
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(parts[0])
		var arg string
		if len(parts) > 1 {
			arg = parts[1]
		}

		// Handle command
		if err := s.handleCommand(ctx, cmd, arg); err != nil {
			return err
		}
	}
}

// handleCommand processes a single pinentry command
func (s *Session) handleCommand(ctx context.Context, cmd, arg string) error {
	switch cmd {
	case "GETPIN":
		return s.handleGetPin(ctx)
	case "SETDESC":
		s.description = UnescapeArg(arg)
		return s.writeOK("")
	case "SETPROMPT":
		s.prompt = UnescapeArg(arg)
		return s.writeOK("")
	case "SETTITLE":
		s.title = UnescapeArg(arg)
		return s.writeOK("")
	case "SETERROR":
		s.error = UnescapeArg(arg)
		return s.writeOK("")
	case "SETKEYINFO":
		s.keyInfo = UnescapeArg(arg)
		return s.writeOK("")
	case "SETOK", "SETCANCEL", "SETNOTOK":
		// Button text (we don't display buttons in terminal mode)
		return s.writeOK("")
	case "SETQUALITYBAR", "SETQUALITYBAR_TT":
		// Quality bar (not supported)
		return s.writeOK("")
	case "GETINFO":
		return s.handleGetInfo(arg)
	case "OPTION":
		// Parse options but ignore them for now
		return s.writeOK("")
	case "CONFIRM", "MESSAGE":
		// Not implemented
		return s.writeOK("")
	case "BYE":
		return s.writeOK("")
	case "RESET":
		s.reset()
		return s.writeOK("")
	default:
		return s.writeError("Unknown command: " + cmd)
	}
}

// handleGetPin retrieves the PIN from ProtonPass
func (s *Session) handleGetPin(ctx context.Context) error {
	// Find matching configuration entry
	itemURI := s.config.FindItemForContext(s.description, s.prompt, s.title, s.keyInfo)

	if DebugMode {
		log.Printf("[DEBUG] Context: desc=%q prompt=%q title=%q keyinfo=%q",
			s.description, s.prompt, s.title, s.keyInfo)
		log.Printf("[DEBUG] Matched item: %s", itemURI)
	}

	if itemURI == "" {
		return s.writeError("No ProtonPass item configured for this context")
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Retrieve password from ProtonPass
	password, err := s.client.RetrievePassword(timeoutCtx, itemURI)
	if err != nil {
		return s.writeError(fmt.Sprintf("Failed to retrieve password: %v", err))
	}

	// Track for cleanup
	s.sensitiveData = append(s.sensitiveData, password)
	defer protonpass.ZeroBytes(password)

	// Send password
	return s.writeData(password)
}

// handleGetInfo handles GETINFO requests
func (s *Session) handleGetInfo(arg string) error {
	switch arg {
	case "version":
		return s.writeData([]byte(Version))
	case "pid":
		return s.writeData([]byte(fmt.Sprintf("%d", os.Getpid())))
	case "flavor":
		return s.writeData([]byte("proton"))
	case "ttyinfo":
		return s.writeData([]byte("terminal"))
	default:
		return s.writeError("Unknown info: " + arg)
	}
}

// reset clears the session state
func (s *Session) reset() {
	s.description = ""
	s.prompt = ""
	s.title = ""
	s.error = ""
	s.keyInfo = ""
}

// writeOK writes an OK response
func (s *Session) writeOK(message string) error {
	if message != "" {
		_, err := fmt.Fprintf(s.writer, "OK %s\n", message)
		return err
	}
	_, err := fmt.Fprintf(s.writer, "OK\n")
	return err
}

// writeError writes an error response
func (s *Session) writeError(message string) error {
	// ERR code message format
	// Using generic error code 83886179 (GPG_ERR_GENERAL)
	_, err := fmt.Fprintf(s.writer, "ERR 83886179 %s\n", EscapeArg(message))
	return err
}

// writeData writes data response
func (s *Session) writeData(data []byte) error {
	// Data must be percent-encoded for special characters
	encoded := PercentEncode(data)
	_, err := fmt.Fprintf(s.writer, "D %s\n", encoded)
	if err != nil {
		return err
	}
	return s.writeOK("")
}

// Cleanup zeros all sensitive data tracked by the session
func (s *Session) Cleanup() {
	for _, data := range s.sensitiveData {
		protonpass.ZeroBytes(data)
	}
	s.sensitiveData = nil
}
