package logger

import (
	"fmt"
	"io"
	"strings"
)

// Global debug mode flag
var debugMode bool

// SetDebugMode sets the global debug mode for logging
func SetDebugMode(debug bool) {
	debugMode = debug
}

// ShouldLog determines if a log prefix should be shown based on debug mode
func ShouldLog(prefix string) bool {
	if debugMode {
		return true // Show all logs in debug mode
	}

	// In non-debug mode, only show proxy and backend logs
	return strings.Contains(prefix, "[proxy]") || strings.Contains(prefix, "[backend]")
}

// Printf prints a formatted log message if the prefix should be logged
func Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if ShouldLog(msg) {
		fmt.Print(msg)
	}
}

// Println prints a log message if the prefix should be logged
func Println(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if ShouldLog(msg) {
		fmt.Println(msg)
	}
}

// PrefixWriter wraps an io.Writer and prefixes each line with a given prefix
type PrefixWriter struct {
	prefix string
	writer io.Writer
	buffer []byte
}

// NewPrefixWriter creates a new PrefixWriter
func NewPrefixWriter(prefix string, writer io.Writer) *PrefixWriter {
	return &PrefixWriter{
		prefix: prefix,
		writer: writer,
		buffer: []byte{},
	}
}

// Write implements io.Writer interface
func (pw *PrefixWriter) Write(p []byte) (n int, err error) {
	// Add to buffer
	pw.buffer = append(pw.buffer, p...)

	// Process complete lines
	lines := strings.Split(string(pw.buffer), "\n")

	// Keep the last incomplete line in buffer
	if len(lines) > 0 && !strings.HasSuffix(string(pw.buffer), "\n") {
		pw.buffer = []byte(lines[len(lines)-1])
		lines = lines[:len(lines)-1]
	} else {
		pw.buffer = []byte{}
	}

	// Write prefixed lines
	for _, line := range lines {
		if line != "" || strings.HasSuffix(string(p), "\n") {
			_, err := fmt.Fprintf(pw.writer, "%s%s\n", pw.prefix, line)
			if err != nil {
				return len(p), err
			}
		}
	}

	return len(p), nil
}
