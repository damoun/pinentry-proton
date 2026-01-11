// Package config handles configuration loading and management for pinentry-proton.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the pinentry-proton configuration
type Config struct {
	// Default item to use if no match is found
	DefaultItem string `yaml:"default_item"`

	// Timeout for password retrieval in seconds (default: 60)
	Timeout int `yaml:"timeout"`

	// Map of context matchers to ProtonPass item URIs
	Mappings []Mapping `yaml:"mappings"`
}

// Mapping represents a context-to-item mapping
type Mapping struct {
	// Name of this mapping (for documentation)
	Name string `yaml:"name"`

	// ProtonPass item URI (pass://vault/item/password)
	Item string `yaml:"item"`

	// Matchers for pinentry context
	Match MatchCriteria `yaml:"match"`
}

// MatchCriteria defines what to match in the pinentry request
type MatchCriteria struct {
	// Match against SETDESC (description)
	Description string `yaml:"description"`

	// Match against SETPROMPT
	Prompt string `yaml:"prompt"`

	// Match against SETTITLE
	Title string `yaml:"title"`

	// Match against SETKEYINFO
	KeyInfo string `yaml:"keyinfo"`
}

// Load loads the configuration from standard locations
func Load() (*Config, error) {
	// Try loading from these locations in order:
	// 1. $PINENTRY_PROTON_CONFIG
	// 2. $XDG_CONFIG_HOME/pinentry-proton/config.yaml
	// 3. $HOME/.config/pinentry-proton/config.yaml
	// 4. $HOME/.pinentry-proton.yaml

	configPaths := getConfigPaths()

	var config *Config
	var err error
	for _, path := range configPaths {
		if _, statErr := os.Stat(path); statErr == nil {
			config, err = loadFromFile(path)
			if err != nil {
				return nil, fmt.Errorf("error loading config from %s: %w", path, err)
			}
			break
		}
	}

	if config == nil {
		// No config file found, use empty config
		config = &Config{
			Timeout: 60,
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// getConfigPaths returns the list of configuration file paths to check
func getConfigPaths() []string {
	configPaths := []string{}

	if envPath := os.Getenv("PINENTRY_PROTON_CONFIG"); envPath != "" {
		configPaths = append(configPaths, envPath)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configPaths = append(configPaths, filepath.Join(xdgConfig, "pinentry-proton", "config.yaml"))
		}
		configPaths = append(configPaths,
			filepath.Join(homeDir, ".config", "pinentry-proton", "config.yaml"),
			filepath.Join(homeDir, ".pinentry-proton.yaml"),
		)
	}

	return configPaths
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: Config file path from user's environment, not external input
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 60
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate default item if provided
	if c.DefaultItem != "" {
		if !strings.HasPrefix(c.DefaultItem, "pass://") {
			return fmt.Errorf("default_item must be a ProtonPass URI (pass://...)")
		}
	}

	// Validate mappings
	for i, mapping := range c.Mappings {
		if err := mapping.Validate(i); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks if a mapping is valid
func (m *Mapping) Validate(index int) error {
	if m.Item == "" {
		return fmt.Errorf("mapping %d (%s): item is required", index, m.Name)
	}
	if !strings.HasPrefix(m.Item, "pass://") {
		return fmt.Errorf("mapping %d (%s): item must be a ProtonPass URI (pass://...)", index, m.Name)
	}
	if m.Match.Description == "" && m.Match.Prompt == "" &&
		m.Match.Title == "" && m.Match.KeyInfo == "" {
		return fmt.Errorf("mapping %d (%s): at least one match criterion is required", index, m.Name)
	}
	return nil
}

// FindItemForContext finds the appropriate ProtonPass item URI for the given context
func (c *Config) FindItemForContext(description, prompt, title, keyInfo string) string {
	// Try to find a matching mapping
	for _, mapping := range c.Mappings {
		if mapping.Matches(description, prompt, title, keyInfo) {
			return mapping.Item
		}
	}

	// Fall back to default item
	return c.DefaultItem
}

// Matches checks if this mapping matches the given context
func (m *Mapping) Matches(description, prompt, title, keyInfo string) bool {
	// All non-empty criteria must match
	if m.Match.Description != "" && !matchesPattern(description, m.Match.Description) {
		return false
	}
	if m.Match.Prompt != "" && !matchesPattern(prompt, m.Match.Prompt) {
		return false
	}
	if m.Match.Title != "" && !matchesPattern(title, m.Match.Title) {
		return false
	}
	if m.Match.KeyInfo != "" && !matchesPattern(keyInfo, m.Match.KeyInfo) {
		return false
	}
	return true
}

// matchesPattern checks if the value matches the pattern
// For now, we use simple substring matching
// This could be extended to support regex or glob patterns
func matchesPattern(value, pattern string) bool {
	if pattern == "*" {
		return true
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(pattern))
}
