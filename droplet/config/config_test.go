package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTryParse_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "test.yaml")

	content := `kcrypt:
  remote_unlock:
    edgevpn_token: "test-token"
    public_key: "test-pub"
    private_key: "test-priv"
    ntp_server: "time.cloudflare.com"
    debug:
      enabled: true
      log_level: -1
      password: "test-password"
      bypass_password_test: true
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	conf := tryParse(configFile)
	if conf == nil {
		t.Fatal("Expected config, got nil")
	}
	if conf.EdgeVPNToken != "test-token" {
		t.Fatalf("Expected 'test-token', got %q", conf.EdgeVPNToken)
	}
	if conf.PublicKey != "test-pub" {
		t.Fatalf("Expected 'test-pub', got %q", conf.PublicKey)
	}
	if conf.PrivateKey != "test-priv" {
		t.Fatalf("Expected 'test-priv', got %q", conf.PrivateKey)
	}
	if !conf.DebugConfig.Enabled {
		t.Fatal("Expected debug enabled")
	}
	if conf.DebugConfig.Password != "test-password" {
		t.Fatalf("Expected 'test-password', got %q", conf.DebugConfig.Password)
	}
	if !conf.DebugConfig.BypassPasswordTest {
		t.Fatal("Expected bypass_password_test=true")
	}
}

func TestTryParse_NonYAML(t *testing.T) {
	conf := tryParse("/tmp/test.txt")
	if conf != nil {
		t.Fatal("Expected nil for non-YAML file")
	}
}

func TestTryParse_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(configFile, []byte("{{{{bad yaml"), 0644); err != nil {
		t.Fatal(err)
	}
	conf := tryParse(configFile)
	if conf != nil {
		t.Fatal("Expected nil for invalid YAML")
	}
}

func TestTryParse_NoKcryptSection(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "nokcrypt.yaml")
	content := `other_section:
  key: value
`
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	conf := tryParse(configFile)
	// tryParse returns a Config with empty fields, not nil
	if conf == nil {
		t.Fatal("Expected non-nil config")
	}
	if conf.EdgeVPNToken != "" {
		t.Fatalf("Expected empty token, got %q", conf.EdgeVPNToken)
	}
}

func TestIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "complete",
			config: Config{
				EdgeVPNToken: "token",
				PublicKey:    "pub",
				PrivateKey:   "priv",
			},
			expected: true,
		},
		{
			name: "missing token",
			config: Config{
				PublicKey:  "pub",
				PrivateKey: "priv",
			},
			expected: false,
		},
		{
			name: "missing public key",
			config: Config{
				EdgeVPNToken: "token",
				PrivateKey:   "priv",
			},
			expected: false,
		},
		{
			name: "missing private key",
			config: Config{
				EdgeVPNToken: "token",
				PublicKey:    "pub",
			},
			expected: false,
		},
		{
			name:     "all empty",
			config:   Config{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsComplete(); got != tt.expected {
				t.Fatalf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFindConfig_MergesMultipleFiles(t *testing.T) {
	dir := t.TempDir()

	// Write first config with token and public key
	config1 := filepath.Join(dir, "01.yaml")
	if err := os.WriteFile(config1, []byte(`kcrypt:
  remote_unlock:
    edgevpn_token: "my-token"
    public_key: "my-pub"
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Write second config with private key
	config2 := filepath.Join(dir, "02.yaml")
	if err := os.WriteFile(config2, []byte(`kcrypt:
  remote_unlock:
    private_key: "my-priv"
    ntp_server: "pool.ntp.org"
`), 0644); err != nil {
		t.Fatal(err)
	}

	conf, err := findConfig([]string{dir})
	if err != nil {
		t.Fatalf("findConfig failed: %v", err)
	}

	if conf.EdgeVPNToken != "my-token" {
		t.Fatalf("Expected 'my-token', got %q", conf.EdgeVPNToken)
	}
	if conf.PublicKey != "my-pub" {
		t.Fatalf("Expected 'my-pub', got %q", conf.PublicKey)
	}
	if conf.PrivateKey != "my-priv" {
		t.Fatalf("Expected 'my-priv', got %q", conf.PrivateKey)
	}
	if conf.NTPServer != "pool.ntp.org" {
		t.Fatalf("Expected 'pool.ntp.org', got %q", conf.NTPServer)
	}
}

func TestFindConfig_FirstValueWins(t *testing.T) {
	dir := t.TempDir()

	config1 := filepath.Join(dir, "01.yaml")
	if err := os.WriteFile(config1, []byte(`kcrypt:
  remote_unlock:
    edgevpn_token: "first-token"
    public_key: "first-pub"
    private_key: "first-priv"
`), 0644); err != nil {
		t.Fatal(err)
	}

	config2 := filepath.Join(dir, "02.yaml")
	if err := os.WriteFile(config2, []byte(`kcrypt:
  remote_unlock:
    edgevpn_token: "second-token"
    public_key: "second-pub"
    private_key: "second-priv"
`), 0644); err != nil {
		t.Fatal(err)
	}

	conf, err := findConfig([]string{dir})
	if err != nil {
		t.Fatalf("findConfig failed: %v", err)
	}

	// First complete config should win (stops after IsComplete)
	if conf.EdgeVPNToken != "first-token" {
		t.Fatalf("Expected first-token, got %q", conf.EdgeVPNToken)
	}
}

func TestFindConfig_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	conf, err := findConfig([]string{dir})
	if err != nil {
		t.Fatalf("findConfig failed: %v", err)
	}
	if conf.IsComplete() {
		t.Fatal("Expected incomplete config from empty dir")
	}
}

func TestFindConfig_HttpPullConfig(t *testing.T) {
	dir := t.TempDir()

	config1 := filepath.Join(dir, "01.yaml")
	if err := os.WriteFile(config1, []byte(`kcrypt:
  remote_unlock:
    edgevpn_token: "token"
    public_key: "pub"
    private_key: "priv"
    http_pull:
      - "192.168.1.100:505"
      - "192.168.1.101:505"
`), 0644); err != nil {
		t.Fatal(err)
	}

	conf, err := findConfig([]string{dir})
	if err != nil {
		t.Fatalf("findConfig failed: %v", err)
	}

	if len(conf.HttpPull) != 2 {
		t.Fatalf("Expected 2 http_pull entries, got %d", len(conf.HttpPull))
	}
}
