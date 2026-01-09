package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
default_item: "pass://Personal/Default/password"
timeout: 30
mappings:
  - name: "Test Mapping"
    item: "pass://Work/Test/password"
    match:
      description: "test"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	
	// Set environment variable to use our temp config
	os.Setenv("PINENTRY_PROTON_CONFIG", configPath)
	defer os.Unsetenv("PINENTRY_PROTON_CONFIG")
	
	config, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if config.DefaultItem != "pass://Personal/Default/password" {
		t.Errorf("DefaultItem = %q, want %q", config.DefaultItem, "pass://Personal/Default/password")
	}
	
	if config.Timeout != 30 {
		t.Errorf("Timeout = %d, want 30", config.Timeout)
	}
	
	if len(config.Mappings) != 1 {
		t.Errorf("Mappings length = %d, want 1", len(config.Mappings))
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config with default",
			config: Config{
				DefaultItem: "pass://Personal/Default/password",
				Timeout:     60,
			},
			wantErr: false,
		},
		{
			name: "Valid config with mappings",
			config: Config{
				Timeout: 60,
				Mappings: []Mapping{
					{
						Name: "Test",
						Item: "pass://Work/Test/password",
						Match: MatchCriteria{
							Description: "test",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid default item URI",
			config: Config{
				DefaultItem: "invalid-uri",
			},
			wantErr: true,
		},
		{
			name: "Invalid mapping item URI",
			config: Config{
				Mappings: []Mapping{
					{
						Name: "Test",
						Item: "invalid-uri",
						Match: MatchCriteria{
							Description: "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Mapping without item",
			config: Config{
				Mappings: []Mapping{
					{
						Name: "Test",
						Match: MatchCriteria{
							Description: "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Mapping without match criteria",
			config: Config{
				Mappings: []Mapping{
					{
						Name: "Test",
						Item: "pass://Work/Test/password",
						Match: MatchCriteria{},
					},
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindItemForContext(t *testing.T) {
	config := &Config{
		DefaultItem: "pass://Personal/Default/password",
		Mappings: []Mapping{
			{
				Name: "GitHub",
				Item: "pass://Work/GitHub/password",
				Match: MatchCriteria{
					Description: "github",
				},
			},
			{
				Name: "GPG Key",
				Item: "pass://Personal/GPG/password",
				Match: MatchCriteria{
					KeyInfo: "ABCD1234",
				},
			},
			{
				Name: "Multiple criteria",
				Item: "pass://Work/Server/password",
				Match: MatchCriteria{
					Description: "production",
					Title:       "ssh",
				},
			},
		},
	}
	
	tests := []struct {
		name        string
		description string
		prompt      string
		title       string
		keyInfo     string
		want        string
	}{
		{
			name:        "Match by description",
			description: "Access GitHub repository",
			want:        "pass://Work/GitHub/password",
		},
		{
			name:    "Match by keyinfo",
			keyInfo: "ABCD1234",
			want:    "pass://Personal/GPG/password",
		},
		{
			name:        "Match multiple criteria",
			description: "production server",
			title:       "SSH Authentication",
			want:        "pass://Work/Server/password",
		},
		{
			name:        "No match, use default",
			description: "unknown context",
			want:        "pass://Personal/Default/password",
		},
		{
			name:        "Case insensitive match",
			description: "GITHUB repo",
			want:        "pass://Work/GitHub/password",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.FindItemForContext(tt.description, tt.prompt, tt.title, tt.keyInfo)
			if got != tt.want {
				t.Errorf("FindItemForContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		value   string
		pattern string
		want    bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", true},
		{"hello world", "HELLO", true},  // case insensitive
		{"hello world", "goodbye", false},
		{"test", "*", true},
		{"", "something", false},
		{"something", "", true},  // empty pattern matches anything
	}
	
	for _, tt := range tests {
		t.Run(tt.value+"_"+tt.pattern, func(t *testing.T) {
			got := matchesPattern(tt.value, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.value, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMappingMatches(t *testing.T) {
	tests := []struct {
		name        string
		mapping     Mapping
		description string
		prompt      string
		title       string
		keyInfo     string
		want        bool
	}{
		{
			name: "Single criterion match",
			mapping: Mapping{
				Match: MatchCriteria{
					Description: "test",
				},
			},
			description: "test description",
			want:        true,
		},
		{
			name: "Single criterion no match",
			mapping: Mapping{
				Match: MatchCriteria{
					Description: "test",
				},
			},
			description: "other description",
			want:        false,
		},
		{
			name: "Multiple criteria all match",
			mapping: Mapping{
				Match: MatchCriteria{
					Description: "prod",
					Title:       "ssh",
				},
			},
			description: "production server",
			title:       "SSH Key",
			want:        true,
		},
		{
			name: "Multiple criteria partial match",
			mapping: Mapping{
				Match: MatchCriteria{
					Description: "prod",
					Title:       "ssh",
				},
			},
			description: "production server",
			title:       "GPG Key",
			want:        false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mapping.Matches(tt.description, tt.prompt, tt.title, tt.keyInfo)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
