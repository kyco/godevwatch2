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
	// Initialize tracker
	tracker := NewTracker(cfg.BuildStatusDir, cfg.DebugMode)

	// Start tracking
	if err := tracker.Start(); err != nil {
		return fmt.Errorf("failed to start build tracking: %w", err)
	}

	// Track build failure if something goes wrong
	var buildErr error
	defer func() {
		if buildErr != nil {
			if err := tracker.Fail(); err != nil {
				fmt.Printf("[build] Warning: failed to mark build as failed: %v\n", err)
			}
		}
	}()

	for _, rule := range cfg.BuildRules {
		fmt.Printf("[build] Running build: %s\n", rule.Name)

		cmd := exec.Command("sh", "-c", rule.Command)
		cmd.Stdout = logger.NewPrefixWriter("[build] ", os.Stdout)
		cmd.Stderr = logger.NewPrefixWriter("[build] ", os.Stderr)

		if err := cmd.Run(); err != nil {
			buildErr = fmt.Errorf("build failed (%s): %w", rule.Name, err)
			return buildErr
		}

		fmt.Printf("[build] ✓ Build completed: %s\n", rule.Name)
	}

	// Mark build as complete
	if err := tracker.Complete(); err != nil {
		return fmt.Errorf("failed to complete build tracking: %w", err)
	}

	return nil
}
