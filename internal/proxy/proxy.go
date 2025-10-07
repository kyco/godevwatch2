package proxy

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyco/godevwatch/internal/build"
	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/process"
)

//go:embed templates/server-down.html
var serverDownPage string

// Start initializes and starts the proxy server
func Start(cfg *config.Config) error {
	// Setup proxy HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, serverDownPage)
	})

	// Start proxy server in background
	addr := fmt.Sprintf(":%d", cfg.ProxyPort)
	server := &http.Server{Addr: addr}

	go func() {
		fmt.Printf("[proxy] \033[32mStarted proxy server on http://localhost%s\033[0m\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[proxy] Server error: %v\n", err)
		}
	}()

	// Run all build rules in order
	fmt.Println()
	if err := build.RunAll(cfg); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println()

	// Start the application
	appCmd, err := process.Start(cfg)
	if err != nil {
		return err
	}
	fmt.Println()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("[proxy] Press Ctrl+C to stop")

	// Wait for termination signal
	<-sigChan

	// Cleanup
	fmt.Println("\n[proxy] Shutting down...")

	// Kill application process
	if appCmd != nil && appCmd.Process != nil {
		fmt.Println("[proxy] Stopping backend application...")
		appCmd.Process.Kill()
	}

	// Remove build status directory
	fmt.Printf("[proxy] Removing build status directory: %s\n", cfg.BuildStatusDir)
	if err := os.RemoveAll(cfg.BuildStatusDir); err != nil {
		fmt.Printf("[proxy] Warning: failed to remove build status directory: %v\n", err)
	}

	fmt.Println("[proxy] Shutdown complete")
	return nil
}
