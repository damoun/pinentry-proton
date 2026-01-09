package platform

import (
	"runtime"
	"testing"
)

// TestInfo verifies Info returns a non-empty platform string
func TestInfo(t *testing.T) {
	info := Info()

	if info == "" {
		t.Error("Info() returned empty string")
	}

	// Verify it matches one of the expected platforms
	validPlatforms := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"unknown": true,
	}

	if !validPlatforms[info] {
		t.Errorf("Info() returned unexpected platform: %q", info)
	}

	// Verify it matches runtime.GOOS for known platforms
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if info != runtime.GOOS {
			t.Errorf("Expected Info() to return %q on %s, got %q", runtime.GOOS, runtime.GOOS, info)
		}
	}
}

// TestSetup_Success verifies Setup completes without error
func TestSetup_Success(t *testing.T) {
	err := Setup()

	if err != nil {
		t.Errorf("Setup() returned error: %v", err)
	}
}

// TestCleanup_Success verifies Cleanup completes without error
func TestCleanup_Success(t *testing.T) {
	// Cleanup doesn't return an error, just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Cleanup() panicked: %v", r)
		}
	}()

	Cleanup()
}

// TestSetup_Idempotent verifies Setup can be called multiple times
func TestSetup_Idempotent(t *testing.T) {
	// Call Setup multiple times
	for i := 0; i < 3; i++ {
		err := Setup()
		if err != nil {
			t.Errorf("Setup() call %d returned error: %v", i+1, err)
		}
	}
}

// TestCleanup_Idempotent verifies Cleanup can be called multiple times
func TestCleanup_Idempotent(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Cleanup() panicked: %v", r)
		}
	}()

	// Call Cleanup multiple times
	for i := 0; i < 3; i++ {
		Cleanup()
	}
}

// TestSetupAndCleanup verifies Setup and Cleanup work together
func TestSetupAndCleanup(t *testing.T) {
	// Setup
	if err := Setup(); err != nil {
		t.Fatalf("Setup() failed: %v", err)
	}

	// Do some work (just verify Info still works)
	info := Info()
	if info == "" {
		t.Error("Info() returned empty after Setup()")
	}

	// Cleanup
	Cleanup()

	// Verify Info still works after cleanup
	info = Info()
	if info == "" {
		t.Error("Info() returned empty after Cleanup()")
	}
}

// TestMultipleCycles verifies multiple Setup/Cleanup cycles
func TestMultipleCycles(t *testing.T) {
	for i := 0; i < 3; i++ {
		t.Run("cycle", func(t *testing.T) {
			if err := Setup(); err != nil {
				t.Errorf("Cycle %d: Setup() failed: %v", i+1, err)
			}

			info := Info()
			if info == "" {
				t.Errorf("Cycle %d: Info() returned empty", i+1)
			}

			Cleanup()
		})
	}
}
