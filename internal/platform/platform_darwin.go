//go:build darwin
// +build darwin

// Package platform provides platform-specific functionality.
package platform

// Info returns platform-specific information
func Info() string {
	return "darwin"
}

// Setup applies platform-specific security settings
func Setup() error {
	// On macOS, we could integrate with Keychain for optional caching
	// but per AGENTS.md, we don't cache passwords by default
	return nil
}

// Cleanup performs platform-specific cleanup
func Cleanup() {
	// macOS-specific cleanup if needed
}
