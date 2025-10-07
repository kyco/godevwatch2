package logger

import (
	"fmt"
	"io"
	"strings"
)

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
