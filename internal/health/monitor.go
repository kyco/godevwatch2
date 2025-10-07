package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/kyco/godevwatch/internal/config"
	"github.com/kyco/godevwatch/internal/logger"
)

// Status represents the current backend status
type Status int

const (
	StatusDown Status = iota
	StatusUp
)

// Monitor manages backend health monitoring and proxy switching
type Monitor struct {
	config            *config.Config
	status            Status
	statusMu          sync.RWMutex
	proxy             *httputil.ReverseProxy
	backendURL        *url.URL
	healthCheckTicker *time.Ticker
	onStatusChange    func(Status)

	// Client connections for auto-reload
	reloadClients   map[chan string]bool
	reloadClientsMu sync.RWMutex
}

// NewMonitor creates a new backend health monitor
func NewMonitor(cfg *config.Config) *Monitor {
	backendURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", cfg.BackendPort),
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	// Customize proxy error handling
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Don't log connection errors - they're expected when backend is down
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Backend temporarily unavailable: %v", err)
	}

	return &Monitor{
		config:        cfg,
		status:        StatusDown,
		proxy:         proxy,
		backendURL:    backendURL,
		reloadClients: make(map[chan string]bool),
	}
}

// Start begins health monitoring
func (m *Monitor) Start(ctx context.Context) {
	// Initial health check
	go m.checkHealth()

	// Start periodic health checks
	m.healthCheckTicker = time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				m.healthCheckTicker.Stop()
				return
			case <-m.healthCheckTicker.C:
				m.checkHealth()
			}
		}
	}()
}

// checkHealth performs a health check on the backend
func (m *Monitor) checkHealth() {
	// Simple TCP connection check (faster than HTTP)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", m.config.BackendPort), 500*time.Millisecond)

	newStatus := StatusDown
	if err == nil {
		conn.Close()
		newStatus = StatusUp
	}

	m.updateStatus(newStatus)
}

// updateStatus updates the backend status and notifies listeners
func (m *Monitor) updateStatus(newStatus Status) {
	m.statusMu.Lock()
	oldStatus := m.status
	m.status = newStatus
	m.statusMu.Unlock()

	// Notify on status change
	if oldStatus != newStatus {
		logger.Printf("[proxy] Backend status changed: %s -> %s\n",
			statusString(oldStatus), statusString(newStatus))

		if m.onStatusChange != nil {
			m.onStatusChange(newStatus)
		}

		// If backend came online, trigger browser reload
		if newStatus == StatusUp && oldStatus == StatusDown {
			m.triggerReload()
		}
	}
}

// GetStatus returns the current backend status
func (m *Monitor) GetStatus() Status {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()
	return m.status
}

// SetStatusChangeCallback sets a callback for status changes
func (m *Monitor) SetStatusChangeCallback(callback func(Status)) {
	m.onStatusChange = callback
}

// GetProxy returns the reverse proxy for the backend
func (m *Monitor) GetProxy() *httputil.ReverseProxy {
	return m.proxy
}

// triggerReload sends reload signal to all connected browser clients
func (m *Monitor) triggerReload() {
	m.reloadClientsMu.RLock()
	defer m.reloadClientsMu.RUnlock()

	logger.Printf("[proxy] Triggering browser reload for %d client(s)\n", len(m.reloadClients))

	for client := range m.reloadClients {
		select {
		case client <- "reload":
		default:
			// Client not ready to receive, skip
		}
	}
}

// AddReloadClient adds a client for auto-reload notifications
func (m *Monitor) AddReloadClient() <-chan string {
	client := make(chan string, 1)

	m.reloadClientsMu.Lock()
	m.reloadClients[client] = true
	m.reloadClientsMu.Unlock()

	return client
}

// RemoveReloadClient removes a client from auto-reload notifications
func (m *Monitor) RemoveReloadClient(client chan string) {
	m.reloadClientsMu.Lock()
	delete(m.reloadClients, client)
	m.reloadClientsMu.Unlock()
}

// ForceReload manually triggers a browser reload
func (m *Monitor) ForceReload() {
	m.triggerReload()
}

// statusString returns a human-readable status string
func statusString(status Status) string {
	switch status {
	case StatusUp:
		return "UP"
	case StatusDown:
		return "DOWN"
	default:
		return "UNKNOWN"
	}
}
