package runner

import (
	"bytes"
	"io"
	"os"
)

// OutputCapture handles capturing and restoring stdout
// Implements OutputManager interface
type OutputCapture struct {
	oldStdout *os.File
	pipeR     *os.File
	pipeW     *os.File
	output    string
	capturing bool
}

// NewOutputCapture creates a new output capture instance
func NewOutputCapture() OutputManager {
	return &OutputCapture{}
}

// Start begins capturing stdout
func (oc *OutputCapture) Start() error {
	if oc.capturing {
		return nil // Already capturing
	}

	// Store the original stdout
	oc.oldStdout = os.Stdout

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	oc.pipeR = r
	oc.pipeW = w

	// Redirect stdout to the pipe
	os.Stdout = w

	oc.capturing = true
	return nil
}

// Stop stops capturing and returns the captured output
func (oc *OutputCapture) Stop() (string, error) {
	if !oc.capturing {
		return "", nil
	}

	// Close the write end of the pipe
	oc.pipeW.Close()

	// Read all output from the pipe
	var buf bytes.Buffer
	_, err := io.Copy(&buf, oc.pipeR)
	if err != nil {
		return "", err
	}

	// Restore original stdout
	os.Stdout = oc.oldStdout

	// Close the read end of the pipe
	oc.pipeR.Close()

	// Store the captured output
	oc.output = buf.String()
	oc.capturing = false

	return oc.output, nil
}

// StartCapture begins capturing stdout (implements OutputManager interface)
func (oc *OutputCapture) StartCapture() {
	oc.Start() // Ignore error for interface compatibility
}

// StopCapture stops capturing and returns output (implements OutputManager interface)
func (oc *OutputCapture) StopCapture() string {
	output, _ := oc.Stop() // Ignore error for interface compatibility
	return output
}

// Write writes data to the captured output (implements OutputManager interface)
func (oc *OutputCapture) Write(data []byte) (int, error) {
	if oc.capturing && oc.pipeW != nil {
		return oc.pipeW.Write(data)
	}
	return len(data), nil
}

// Capture returns the captured output as bytes (implements OutputManager interface)
func (oc *OutputCapture) Capture() ([]byte, error) {
	return []byte(oc.output), nil
}

