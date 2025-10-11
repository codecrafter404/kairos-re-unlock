package droplet

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os/exec"
	"strings"
)

type Config struct {
	EdgeVPNToken string `yaml:"edgevpn_token"`
	PublicKey    string `yaml:"public_key"`
	PrivateKey   string `yaml:"private_key"`
}

func UnmarshalConfig() (Config, error) {
	cmd := exec.Command("kairos-agent", "config", "get", "kcrypt.remote_unlock")
	var res strings.Builder
	cmd.Stdout = &res
	err := cmd.Run()
	if err != nil {
		return Config{}, fmt.Errorf("Failed to get config using kairos-agent: %w", err)
	}

	output := res.String()
	if output == "" {
		return Config{}, fmt.Errorf("kcrypt.remote_unlock is not set")
	}
	var result Config
	err = yaml.Unmarshal([]byte(output), &result)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to unmarshal: %w", err)
	}

	return result, nil
}
