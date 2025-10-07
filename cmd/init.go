package cmd

import (
	"fmt"
	"os"

	"github.com/kyco/godevwatch/internal/config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new godevwatch.yaml configuration file",
	Long:  `Creates a godevwatch.yaml file in the current directory with default settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config already exists
		if _, err := os.Stat("godevwatch.yaml"); err == nil {
			// Prompt user for confirmation with interactive select
			prompt := promptui.Select{
				Label: "godevwatch.yaml already exists. Overwrite?",
				Items: []string{"Yes", "No"},
				CursorPos: 0, // Default to "Yes"
			}

			_, result, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("prompt failed: %w", err)
			}

			if result != "Yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		// Create default config
		if err := config.Init(); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}

		fmt.Println("✓ Created godevwatch.yaml")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
