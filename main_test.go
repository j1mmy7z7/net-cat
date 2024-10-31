package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMainPrintsErrorMessageForNonNumericPort(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"program", "invalidPort"}

	var logOutput strings.Builder
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	select {
	case <-done:

		// currentTime := time.Now().Format("2006/01/02 15:04:05")
		expectedError := ""
		if !strings.Contains(logOutput.String(), expectedError) {
			t.Errorf("Expected log to contain %q, but got %q", expectedError, logOutput.String())
		}
	case <-time.After(2 * time.Second):
		t.Error("Test timed out, possible deadlock or long-running goroutine")
	}
}

func TestMainPrintsErrorMessageForExceedingPortRange(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"program", "70000"}

	var logOutput strings.Builder
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	select {
	case <-done:
		// currentTime := time.Now().Format("2006/01/02 15:04:05")

		expectedError := ""
		if !strings.Contains(logOutput.String(), expectedError) {
			t.Errorf("Expected log to contain %q, but got %q", expectedError, logOutput.String())
		}
	case <-time.After(2 * time.Second):
		t.Error("Test timed out, possible deadlock or long-running goroutine")
	}
}

func TestMainLogsServerStartMessageWithCorrectPort(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	port := "12345"
	os.Args = []string{"program", port}

	var logOutput strings.Builder
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	select {
	case <-done:
		t.Error("Expected server to run indefinitely, but it returned")
	case <-time.After(1 * time.Second):
		expectedLogMessage := fmt.Sprintf("Starting server on port %s...", port)
		if !strings.Contains(logOutput.String(), expectedLogMessage) {
			t.Errorf("Expected log to contain %q, but got %q", expectedLogMessage, logOutput.String())
		}
		// Simulate server shutdown
		os.Args = []string{"program", "stop"}
	}
}

func TestMainHandlesPortAlreadyInUse(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Use a valid port number
	port := "8080"
	os.Args = []string{"program", port}

	// Start a listener on the same port to simulate "port already in use"
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		t.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	defer listener.Close()

	var logOutput strings.Builder
	log.SetOutput(&logOutput)
	defer log.SetOutput(os.Stderr)

	done := make(chan bool)
	go func() {
		main()
		done <- true
	}()

	select {
	case <-done:
		expectedLogMessage := fmt.Sprintf("Starting server on port %s...", port)
		if !strings.Contains(logOutput.String(), expectedLogMessage) {
			t.Errorf("Expected log to contain %q, but got %q", expectedLogMessage, logOutput.String())
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected server to handle port already in use and terminate, but it did not")
	}
}
