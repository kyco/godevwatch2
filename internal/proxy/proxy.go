package proxy

import (
	_ "embed"
	"fmt"
	"net/http"

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

	// Ensure application process is killed when proxy exits
	defer func() {
		if appCmd != nil && appCmd.Process != nil {
			appCmd.Process.Kill()
		}
	}()

	fmt.Println("[proxy] Press Ctrl+C to stop")

	// Block forever (until Ctrl+C)
	select {}
}
