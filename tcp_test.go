package main

import (
	"net"
	"testing"
)

// Prompts the client to enter their name and reads the input correctly
func TestWelcomePromptsClientName(t *testing.T) {
	server := &Server{
		chat:  make(map[string]net.Conn),
		msgch: make(chan client),
	}
	
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		// if err != nil {
			// t.Fatalf("Failed to accept connection: %v", err)
		// }
		defer conn.Close()

		_, err = Welcome(conn, server)
		if err != nil {
			t.Errorf("Welcome returned an error: %v", err)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	response := make([]byte, 1024)
	_, err = conn.Read(response)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

}
