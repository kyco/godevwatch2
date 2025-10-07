package process

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/logger"
)

// Start executes the run command and keeps it running in the background
func Start(cfg *config.Config) (*exec.Cmd, error) {
	fmt.Printf("[backend] Starting application: %s\n", cfg.RunCmd)

	cmd := exec.Command("sh", "-c", cfg.RunCmd)
	cmd.Stdout = logger.NewPrefixWriter("[backend] ", os.Stdout)
	cmd.Stderr = logger.NewPrefixWriter("[backend] ", os.Stderr)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start application: %w", err)
	}

	fmt.Printf("[backend] ✓ Application started (PID: %d)\n", cmd.Process.Pid)

	return cmd, nil
}
