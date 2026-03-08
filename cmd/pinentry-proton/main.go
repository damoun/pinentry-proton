// Pinentry-Proton: A secure pinentry program that integrates with ProtonPass
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/damoun/pinentry-proton/internal/config"
	"github.com/damoun/pinentry-proton/internal/protocol"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup debug mode
	if protocol.DebugMode {
		log.SetOutput(os.Stderr)
		log.Printf("[DEBUG] Pinentry-Proton v%s starting in debug mode", protocol.Version)
		log.Printf("[DEBUG] Platform: %s", runtime.GOOS)
		log.Printf("[DEBUG] Loaded %d mappings", len(cfg.Mappings))
	} else {
		// Disable logging in production
		log.SetOutput(io.Discard)
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create session
	session := protocol.NewSession(os.Stdin, os.Stdout, cfg)
	defer session.Cleanup()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		if protocol.DebugMode {
			log.Printf("[DEBUG] Received signal: %v, cleaning up...", sig)
		}
		session.Cleanup()
		cancel()
	}()

	// Run session
	if err := session.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		session.Cleanup()
		os.Exit(1)
	}
}
