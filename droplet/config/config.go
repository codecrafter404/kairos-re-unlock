package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
<<<<<<< Updated upstream
	EdgeVPNToken string      `yaml:"edgevpn_token"`
	PublicKey    string      `yaml:"public_key"`
	PrivateKey   string      `yaml:"private_key"`
	DebugConfig  DebugConfig `yaml:"debug"`
	NTPServer    string      `yaml:"ntp_server"`
=======
	EdgeVPNToken   string      `yaml:"edgevpn_token"`
	PublicKey      string      `yaml:"public_key"`
	PrivateKey     string      `yaml:"private_key"`
	DebugConfig    DebugConfig `yaml:"debug"`
	NTPServer      string      `yaml:"ntp_server"`
	DiscordWebhook string      `yaml:"discord_webhook"`
	HttpPull       []string    `yaml:"http_pull"`
>>>>>>> Stashed changes
}

type DebugConfig struct {
	Enabled  bool   `yaml:"enabled"`
	LogLevel int    `yaml:"log_level"`
	Password string `yaml:"password"`
}

func (c Config) IsDebugEnabled() bool {
	// if c.DebugConfig != nil {
	// if c.DebugConfig.Enabled != nil {
	return c.DebugConfig.Enabled
	// }
	// }
}

func (c Config) IsComplete() bool {
	return c.EdgeVPNToken != "" && c.PublicKey != "" && c.PrivateKey != ""
}

type ConfigContainer struct {
	Kcrypt struct {
		Config Config `yaml:"remote_unlock"`
	}
}

var configDirs = []string{"/oem", "/sysroot/oem", "/tmp/oem"}

func UnmarshalConfig() (Config, error) {
	conf, err := findConfig(configDirs)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to find config: %w", err)
	}
	if !conf.IsComplete() {
		return Config{}, fmt.Errorf("Config is not compleate")
	}
	return conf, nil
}

func findConfig(dirs []string) (Config, error) {
	confs := []Config{}
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			// try parse
			if conf := tryParse(path); conf != nil {
				confs = append(confs, *conf)
			}

			return nil
		})
	}

	// if ther're multiple configs found, merge them
	res := Config{}

	for _, c := range confs {
		if res.EdgeVPNToken == "" && c.EdgeVPNToken != "" {
			res.EdgeVPNToken = c.EdgeVPNToken
		}
		if res.PrivateKey == "" && c.PrivateKey != "" {
			res.PrivateKey = c.PrivateKey
		}
		if res.PublicKey == "" && c.PublicKey != "" {
			res.PublicKey = c.PublicKey
		}
<<<<<<< Updated upstream
=======

		if len(res.HttpPull) == 0 && len(c.HttpPull) > 0 {
			res.HttpPull = c.HttpPull
		}

		if res.DiscordWebhook == "" && c.DiscordWebhook != "" {
			res.DiscordWebhook = c.DiscordWebhook
		}
>>>>>>> Stashed changes
		if c.DebugConfig.Enabled && !res.DebugConfig.Enabled {
			res.DebugConfig.Enabled = true
		}
		if c.DebugConfig.LogLevel != 0 && res.DebugConfig.LogLevel == 0 {
			res.DebugConfig.LogLevel = c.DebugConfig.LogLevel
		}
		if c.DebugConfig.Password != "" && res.DebugConfig.Password == "" {
			res.DebugConfig.Password = c.DebugConfig.Password
		}
		if c.NTPServer != "" && res.NTPServer == "" {
			res.NTPServer = c.NTPServer
		}
		if res.IsComplete() {
			break
		}
	}

	return res, nil
}

func tryParse(file string) *Config {
	if !strings.HasSuffix(file, ".yaml") && !strings.HasSuffix(file, ".yml") {
		return nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	var container ConfigContainer

	err = yaml.Unmarshal(data, &container)
	if err != nil {
		return nil
	}

	return &container.Kcrypt.Config
}
