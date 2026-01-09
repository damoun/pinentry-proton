//go:build linux
// +build linux

// Package platform provides platform-specific functionality.
package platform

// Info returns platform-specific information
func Info() string {
	return "linux"
}

// Setup applies platform-specific security settings
func Setup() error {
	// On Linux, we could integrate with libsecret/GNOME Keyring
	// but per AGENTS.md, we don't cache passwords by default
	return nil
}

// Cleanup performs platform-specific cleanup
func Cleanup() {
	// Linux-specific cleanup if needed
}
