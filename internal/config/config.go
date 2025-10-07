package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProxyPort int `yaml:"proxy_port"`
}

const defaultConfigContent = `# godevwatch configuration
proxy_port: 3000
`

// Init creates a new godevwatch.yaml file with default settings
func Init() error {
	return os.WriteFile("godevwatch.yaml", []byte(defaultConfigContent), 0644)
}

// Load reads and parses the godevwatch.yaml configuration file
func Load() (*Config, error) {
	// Check if config file exists
	data, err := os.ReadFile("godevwatch.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("godevwatch.yaml not found. Run 'godevwatch init' to create one")
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults if not specified
	if cfg.ProxyPort == 0 {
		cfg.ProxyPort = 3000
	}

	return &cfg, nil
}
