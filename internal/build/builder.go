package build

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/logger"
)

// RunAll executes all build rules in order
func RunAll(cfg *config.Config) error {
	for _, rule := range cfg.BuildRules {
		fmt.Printf("[build] Running build: %s\n", rule.Name)

		cmd := exec.Command("sh", "-c", rule.Command)
		cmd.Stdout = logger.NewPrefixWriter("[build] ", os.Stdout)
		cmd.Stderr = logger.NewPrefixWriter("[build] ", os.Stderr)

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build failed (%s): %w", rule.Name, err)
		}

		fmt.Printf("[build] ✓ Build completed: %s\n", rule.Name)
	}

	return nil
}
