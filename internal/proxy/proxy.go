package proxy

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/kyco/godevwatch/internal/config"
)

//go:embed templates/server-down.html
var serverDownPage string

// Start initializes and starts the proxy server
func Start(cfg *config.Config) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, serverDownPage)
	})

	addr := fmt.Sprintf(":%d", cfg.ProxyPort)

	// Print success message in green
	fmt.Printf("\033[32mStarted proxy server on http://localhost%s\033[0m\n", addr)
	fmt.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
