package clogger

import (
	"io"
	"log"
)

// consoleWriter is a simple console
// writer.
type consoleWriter struct {
}

var _ io.Writer = (*consoleWriter)(nil)

func (c *consoleWriter) Write(p []byte) (n int, err error) {
	log.Println(string(p))

	return len(p), nil
}
