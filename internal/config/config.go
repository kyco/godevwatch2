package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type BuildRule struct {
	Name    string   `yaml:"name"`
	Watch   []string `yaml:"watch"`
	Ignore  []string `yaml:"ignore,omitempty"`
	Command string   `yaml:"command"`
}

type Config struct {
	ProxyPort      int         `yaml:"proxy_port"`
	BackendPort    int         `yaml:"backend_port"`
	BuildStatusDir string      `yaml:"build_status_dir"`
	BuildRules     []BuildRule `yaml:"build_rules"`
	RunCmd         string      `yaml:"run_cmd"`
	DebugMode      bool        // Set via --debug flag, not from YAML
}

const defaultConfigContent = `# godevwatch configuration file
# Place this file in your project root as godevwatch.yaml

# Port for the development proxy server
proxy_port: 3000

# Port of your backend Go server
backend_port: 8080

# Directory where build status files are stored
build_status_dir: tmp/.build-status

# Build rules define conditional build steps based on file changes
# Rules are executed in order, and only run when matching files change
build_rules:
  - name: "go-build"
    watch:
      - "**/*.go"
    ignore:
      - "**/*_test.go"
      - "vendor/**"
      - "node_modules/**"
    command: "go build -o ./tmp/main ."

# Command to run your application after successful build
run_cmd: "./tmp/main"
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
	if cfg.BackendPort == 0 {
		cfg.BackendPort = 8080
	}
	if cfg.BuildStatusDir == "" {
		cfg.BuildStatusDir = "tmp/.build-status"
	}
	if cfg.RunCmd == "" {
		cfg.RunCmd = "./tmp/main"
	}

	return &cfg, nil
}
