package droplet

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	EdgeVPNToken string `yaml:"edgevpn_token"`
	PublicKey    string `yaml:"public_key"`
	PrivateKey   string `yaml:"private_key"`
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
