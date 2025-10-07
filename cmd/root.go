package cmd

import (
	"fmt"

	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/proxy"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "godevwatch",
	Short: "A development proxy tool",
	Long:  `godevwatch is a CLI tool that starts a proxy server for development purposes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Start proxy server
		return proxy.Start(cfg)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = version
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
}
