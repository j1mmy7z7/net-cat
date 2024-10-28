    // Log file path is invalid or inaccessible
	package main

	import (
	  "os"
	  "testing"
	)
	
	
	    // No error occurs when opening the log file
func TestNoErrorOpeningLogFile(t *testing.T) {
    // Call the Init function to open the log file
    Init()

    // Check if the log file was opened successfully
    if _, err := os.Stat("chat.log"); err != nil {
        t.Fatalf("Error opening log file: %v", err)
    }
}

    // Log message is an empty string

	    // Log file is closed or inaccessible during writing
func TestLogFileClosedOrInaccessibleDuringWriting(t *testing.T) {
    logFile = nil

    logMessage("Test message")

    // Verify no error is thrown
}