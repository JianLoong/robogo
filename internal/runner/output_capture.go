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
	// Add real-time display support
	realTimeDisplay bool
	teeFile         *os.File
	teeDone         chan struct{}
}

// NewOutputCapture creates a new output capture instance
func NewOutputCapture() OutputManager {
	return &OutputCapture{}
}

// Start begins capturing stdout
func (oc *OutputCapture) Start() error {
	return oc.StartWithRealTimeDisplay(true) // Enable real-time display by default
}

// StartWithRealTimeDisplay begins capturing stdout with optional real-time display
func (oc *OutputCapture) StartWithRealTimeDisplay(realTime bool) error {
	if oc.capturing {
		return nil // Already capturing
	}

	// Store the original stdout
	oc.oldStdout = os.Stdout
	oc.realTimeDisplay = realTime

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	oc.pipeR = r
	oc.pipeW = w

	// If real-time display is enabled, create a tee-like setup
	if realTime {
		// Create a temporary file to act as our new stdout
		tempFile, err := os.CreateTemp("", "robogo_stdout_*")
		if err != nil {
			return err
		}
		oc.teeFile = tempFile
		oc.teeDone = make(chan struct{})

		// Start a goroutine to copy from temp file to both pipe and original stdout
		go func() {
			defer tempFile.Close()
			defer close(oc.teeDone)

			// Create a multi-writer for both destinations
			multiWriter := io.MultiWriter(w, oc.oldStdout)
			io.Copy(multiWriter, tempFile)
		}()

		// Redirect stdout to our temp file
		os.Stdout = tempFile
	} else {
		// Traditional capture - redirect stdout to the pipe only
		os.Stdout = w
	}

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

	// If using real-time display, close the tee file and wait for completion
	if oc.realTimeDisplay && oc.teeFile != nil {
		oc.teeFile.Close()
		<-oc.teeDone // Wait for the tee goroutine to finish
	}

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
	oc.realTimeDisplay = false
	oc.teeFile = nil
	oc.teeDone = nil

	return oc.output, nil
}

// StartCapture begins capturing stdout
func (oc *OutputCapture) StartCapture() {
	_ = oc.Start() // Start capture, error handling done in Start()
}

// StopCapture stops capturing and returns output
func (oc *OutputCapture) StopCapture() string {
	output, _ := oc.Stop() // Error handling done in Stop()
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
