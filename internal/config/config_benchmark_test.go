package config

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkLoad benchmarks config loading from file
func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
default_item: "pass://vault/item/password"
timeout: 60

mappings:
  - name: "GPG Key"
    item: "pass://gpg/key/password"
    match:
      description: "gpg"
  - name: "SSH Key"
    item: "pass://ssh/key/password"
    match:
      description: "ssh"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	if err := os.Setenv("PINENTRY_PROTON_CONFIG", configPath); err != nil {
		b.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("PINENTRY_PROTON_CONFIG"); err != nil {
			b.Errorf("Failed to unset environment variable: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load()
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

// BenchmarkLoad_LargeConfig benchmarks loading config with many mappings
func BenchmarkLoad_LargeConfig(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with 100 mappings
	config := `default_item: "pass://vault/item/password"
timeout: 60
mappings:
`
	for i := 0; i < 100; i++ {
		config += `  - name: "Mapping ` + string(rune('0'+i%10)) + `"
    item: "pass://vault/item` + string(rune('0'+i%10)) + `/password"
    match:
      description: "test` + string(rune('0'+i%10)) + `"
`
	}

	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	if err := os.Setenv("PINENTRY_PROTON_CONFIG", configPath); err != nil {
		b.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("PINENTRY_PROTON_CONFIG"); err != nil {
			b.Errorf("Failed to unset environment variable: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load()
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

// BenchmarkFindItemForContext benchmarks context matching
func BenchmarkFindItemForContext(b *testing.B) {
	config := &Config{
		DefaultItem: "pass://default/item/password",
		Timeout:     60,
		Mappings: []Mapping{
			{
				Name: "GPG Key",
				Item: "pass://gpg/key/password",
				Match: MatchCriteria{
					Description: "gpg",
				},
			},
			{
				Name: "SSH Key",
				Item: "pass://ssh/key/password",
				Match: MatchCriteria{
					Description: "ssh",
				},
			},
			{
				Name: "GitHub",
				Item: "pass://work/github/password",
				Match: MatchCriteria{
					Description: "github",
				},
			},
		},
	}

	contexts := []struct {
		desc    string
		prompt  string
		title   string
		keyInfo string
	}{
		{"gpg: signing key ABCD1234", "Passphrase", "GPG", "ABCD1234"},
		{"ssh key ~/.ssh/id_ed25519", "Passphrase:", "SSH", ""},
		{"Enter password for github.com", "Password:", "GitHub", ""},
		{"unknown context", "PIN:", "Unknown", ""},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := contexts[i%len(contexts)]
		_ = config.FindItemForContext(ctx.desc, ctx.prompt, ctx.title, ctx.keyInfo)
	}
}

// BenchmarkFindItemForContext_ManyMappings benchmarks with 100 mappings
func BenchmarkFindItemForContext_ManyMappings(b *testing.B) {
	config := &Config{
		DefaultItem: "pass://default/item/password",
		Timeout:     60,
		Mappings:    make([]Mapping, 100),
	}

	// Create 100 mappings
	for i := 0; i < 100; i++ {
		config.Mappings[i] = Mapping{
			Name: "Mapping " + string(rune('0'+i%10)),
			Item: "pass://vault/item" + string(rune('0'+i%10)) + "/password",
			Match: MatchCriteria{
				Description: "test" + string(rune('0'+i%10)),
			},
		}
	}

	// Test matching at different positions
	tests := []struct {
		name string
		desc string
	}{
		{"FirstMatch", "test0"},
		{"MiddleMatch", "test5"},
		{"LastMatch", "test9"},
		{"NoMatch", "nomatch"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = config.FindItemForContext(tt.desc, "", "", "")
			}
		})
	}
}

// BenchmarkMatchesPattern benchmarks pattern matching
func BenchmarkMatchesPattern(b *testing.B) {
	tests := []struct {
		name    string
		text    string
		pattern string
	}{
		{"ExactMatch", "hello world", "hello world"},
		{"SubstringMatch", "hello world test", "world"},
		{"CaseInsensitive", "Hello World", "hello"},
		{"NoMatch", "hello world", "goodbye"},
		{"Wildcard", "anything", "*"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = matchesPattern(tt.text, tt.pattern)
			}
		})
	}
}

// BenchmarkMappingMatches benchmarks full mapping match logic
func BenchmarkMappingMatches(b *testing.B) {
	mapping := &Mapping{
		Name: "Test Mapping",
		Item: "pass://test/item/password",
		Match: MatchCriteria{
			Description: "gpg",
			Title:       "Passphrase",
		},
	}

	tests := []struct {
		name    string
		desc    string
		prompt  string
		title   string
		keyInfo string
	}{
		{"BothMatch", "gpg: signing key", "PIN:", "Passphrase", ""},
		{"OneMatch", "gpg: signing key", "PIN:", "Unknown", ""},
		{"NoMatch", "ssh key", "PIN:", "Unknown", ""},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = mapping.Matches(tt.desc, tt.prompt, tt.title, tt.keyInfo)
			}
		})
	}
}

// BenchmarkValidate benchmarks config validation
func BenchmarkValidate(b *testing.B) {
	config := &Config{
		DefaultItem: "pass://default/item/password",
		Timeout:     60,
		Mappings: []Mapping{
			{
				Name: "GPG Key",
				Item: "pass://gpg/key/password",
				Match: MatchCriteria{
					Description: "gpg",
				},
			},
			{
				Name: "SSH Key",
				Item: "pass://ssh/key/password",
				Match: MatchCriteria{
					Description: "ssh",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
