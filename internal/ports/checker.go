package ports

import (
	"fmt"
	"net"
	"time"
)

// IsAvailable checks if a port is available for use
func IsAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// WaitForAvailable waits for a port to become available (used after starting a server)
func WaitForAvailable(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !IsAvailable(port) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for port %d", port)
}
