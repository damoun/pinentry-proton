// +build !darwin,!linux

// Package platform provides platform-specific functionality.
package platform

// Info returns platform-specific information
func Info() string {
	return "unsupported"
}

// Setup applies platform-specific security settings
func Setup() error {
	return nil
}

// Cleanup performs platform-specific cleanup
func Cleanup() {
	// No cleanup needed
}
