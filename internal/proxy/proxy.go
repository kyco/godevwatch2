package proxy

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/kyco/godevwatch/internal/build"
	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/health"
	"github.com/kyco/godevwatch/internal/process"
	"github.com/kyco/godevwatch/internal/watcher"
)

//go:embed templates/server-down.html
var serverDownPage string

// Start initializes and starts the proxy server
func Start(cfg *config.Config) error {
	// Create health monitor
	monitor := health.NewMonitor(cfg)

	// Setup proxy HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if monitor.GetStatus() == health.StatusUp {
			// Backend is up, proxy the request
			monitor.GetProxy().ServeHTTP(w, r)
		} else {
			// Backend is down, show waiting page
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, serverDownPage)
		}
	})

	// Health check endpoint
	http.HandleFunc("/__health", func(w http.ResponseWriter, r *http.Request) {
		if monitor.GetStatus() == health.StatusUp {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "Backend Down")
		}
	})

	// Server-Sent Events endpoint for auto-reload
	http.HandleFunc("/__reload", func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Get reload client channel
		clientChan := monitor.AddReloadClient()
		defer func() {
			// Close cleanup is handled by the monitor when connection ends
		}()

		// Keep connection alive and wait for reload signal
		for {
			select {
			case msg := <-clientChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
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

	// Start health monitor
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	defer monitorCancel()
	monitor.Start(monitorCtx)

	// Run initial build for all rules (don't crash on failure)
	fmt.Println()
	var appCmd *exec.Cmd
	if err := build.RunAll(cfg); err != nil {
		fmt.Printf("[proxy] \033[31mInitial build failed: %v\033[0m\n", err)
		fmt.Printf("[proxy] \033[33mProxy will continue running. Fix the build errors and file watcher will rebuild automatically.\033[0m\n")
	} else {
		fmt.Printf("[proxy] \033[32mInitial build completed successfully\033[0m\n")

		// Only try to start the application if build succeeded
		var err error
		appCmd, err = process.Start(cfg)
		if err != nil {
			fmt.Printf("[proxy] \033[31mFailed to start backend: %v\033[0m\n", err)
			fmt.Printf("[proxy] \033[33mProxy will continue running. Backend will start after successful build.\033[0m\n")
		}
	}
	fmt.Println()

	// Create and start file watcher with backend restart capability
	w, err := watcher.NewWatcher(cfg)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Set up watcher to restart backend and trigger reload on successful builds
	w.SetBuildSuccessCallback(func() {
		fmt.Printf("[proxy] Build succeeded, starting/restarting backend...\n")

		// Kill existing backend if running
		if appCmd != nil && appCmd.Process != nil {
			fmt.Printf("[proxy] Stopping existing backend...\n")
			appCmd.Process.Kill()
			appCmd.Wait() // Wait for process to exit
		}

		// Start new backend
		newCmd, err := process.Start(cfg)
		if err != nil {
			fmt.Printf("[proxy] \033[31mFailed to start backend: %v\033[0m\n", err)
		} else {
			appCmd = newCmd
			fmt.Printf("[proxy] \033[32mBackend started successfully\033[0m\n")
			// Monitor will detect the new backend and trigger reload automatically
		}
	})

	// Start watcher in background
	ctx, cancel := context.WithCancel(context.Background())
	watcherDone := make(chan error, 1)
	go func() {
		watcherDone <- w.Start(ctx)
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("[proxy] Press Ctrl+C to stop")

	// Wait for termination signal or watcher error
	select {
	case <-sigChan:
		// User requested shutdown
	case err := <-watcherDone:
		if err != nil {
			fmt.Printf("[proxy] Watcher error: %v\n", err)
		}
	}

	// Cancel watcher context
	cancel()

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
