package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Simple HTTP server for testing godevwatch
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Example Backend</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 600px; margin: 0 auto; }
        .status { background: #e8f5e8; padding: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Example Backend Server</h1>
        <div class="status">
            <h2>Status: Running</h2>
            <p><strong>Time:</strong> %s</p>
            <p><strong>Port:</strong> 8080</p>
            <p>This is a simple example backend for testing godevwatch.</p>
        </div>
        <h3>Features:</h3>
        <ul>
            <li>✅ HTTP server on port 8080</li>
            <li>✅ Auto-reload when files change</li>
            <li>✅ Build tracking and monitoring</li>
        </ul>
    </div>
</body>
</html>`

		fmt.Fprintf(w, html, time.Now().Format("2006-01-02 15:04:05"))
	})

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	fmt.Println("🚀 Starting example backend server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
