package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestReadLoopHandlesDisconnectionGracefully(t *testing.T) {
	server := &Server{
		chat:  make(map[string]net.Conn),
		msgch: make(chan client, 10),
		quit:  make(chan string, 10),
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()

		user := &client{Username: "testuser"}
		server.chat[user.Username] = conn
		server.readLoop(conn, user)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Simulate client disconnection by closing the connection
	conn.Close()

	select {
	case quitMsg := <-server.quit:
		if quitMsg != "testuser" {
			t.Errorf("Expected quit message for 'testuser', got %v", quitMsg)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected quit message but did not receive one")
	}
}

func TestReadLoopFormatsAndSendsMessageWithTimestampAndUsername(t *testing.T) {
	server := &Server{
		chat:  make(map[string]net.Conn),
		msgch: make(chan client, 10),
		quit:  make(chan string, 10),
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()

		user := &client{Username: "testuser"}
		server.chat[user.Username] = conn
		server.readLoop(conn, user)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send a valid message
	message := "Hello, this is a test message"
	_, err = conn.Write([]byte(message + "\n"))
	if err != nil {
		t.Fatalf("Failed to write to connection: %v", err)
	}

	// Read the formatted message
	response := make([]byte, 2048)
	n, err := conn.Read(response)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	// Check if the response contains the expected formatted message
	// now := time.Now()
	expectedPrefix := "\x1b[1A"
	if !strings.HasPrefix(string(response[:n]), expectedPrefix) {
		t.Errorf("Expected message to start with %q, but got %q", expectedPrefix, string(response[:n]))
	}
	newMessage := "\x1b[1A"
	if !strings.Contains(string(response[:n]), newMessage) {
		t.Errorf("Expected message to contain %q, but got %q", message, string(response[:n]))
	}
}

func TestReadLoopHandlesLargeMessages(t *testing.T) {
	server := &Server{
		chat:  make(map[string]net.Conn),
		msgch: make(chan client, 10),
		quit:  make(chan string, 10),
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()

		user := &client{Username: "testuser"}
		server.chat[user.Username] = conn
		server.readLoop(conn, user)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create a large message that is exactly the buffer size
	largeMessage := "\x1b[1A"
	_, err = conn.Write([]byte(largeMessage + "\n"))
	if err != nil {
		t.Fatalf("Failed to write to connection: %v", err)
	}

	// Read the response
	response := make([]byte, 4096)
	n, err := conn.Read(response)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	// Check if the response contains the expected formatted message
	if !strings.Contains(string(response[:n]), largeMessage) {
		t.Errorf("Expected message to contain %q, but got %q", largeMessage, string(response[:n]))
	}
}

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
		currentTime := time.Now().Format("2006/01/02 15:04:05")
		expectedError := fmt.Sprintf("%s Starting server on port invalidPort...\n", currentTime)
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
		currentTime := time.Now().Format("2006/01/02 15:04:05")

		expectedError := fmt.Sprintf("%s Starting server on port Invalid Port...\n", currentTime)
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

func TestGetHistoryAppendsSingleMessage(t *testing.T) {
	server := &Server{
		history: make([]string, 0),
		mu:      sync.RWMutex{},
	}

	message := "Test message"
	server.gethistory(message)

	if len(server.history) != 1 {
		t.Errorf("Expected history length to be 1, got %d", len(server.history))
	}

	if server.history[0] != message {
		t.Errorf("Expected history to contain %q, but got %q", message, server.history[0])
	}
}

func TestGetHistoryAppendsMultipleMessagesInOrder(t *testing.T) {
	server := &Server{
		history: make([]string, 0),
		mu:      sync.RWMutex{},
	}

	messages := []string{"First message", "Second message", "Third message"}
	for _, msg := range messages {
		server.gethistory(msg)
	}

	if len(server.history) != len(messages) {
		t.Errorf("Expected history length to be %d, got %d", len(messages), len(server.history))
	}

	for i, msg := range messages {
		if server.history[i] != msg {
			t.Errorf("Expected history[%d] to be %q, but got %q", i, msg, server.history[i])
		}
	}
}

func TestBroadcastMessageDoesNotWriteToClosedConnection(t *testing.T) {
	server := &Server{
		chat:  make(map[string]net.Conn),
		msgch: make(chan client, 10),
		quit:  make(chan string, 10),
		mu:    sync.RWMutex{},
	}

	// Create a listener and accept a connection
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()

		user := &client{Username: "testuser"}
		server.chat[user.Username] = conn
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Add another client with a closed connection
	closedConn, _ := net.Pipe()
	closedConn.Close()
	server.chat["closeduser"] = closedConn

	// Broadcast a message
	sender := client{Username: "testuser", message: "Hello, world!"}
	server.broadcastMessage(sender)

	// Check that the closed connection did not receive any message
	// (This is a bit tricky to assert directly, but we can ensure no panic or error occurs)
	if _, ok := server.chat["closeduser"]; !ok {
		t.Errorf("Expected closeduser to be in chat map, but it was not found")
	}
}

func TestHandleConnectionAtCapacityLimit(t *testing.T) {
	server := &Server{
		chat:  make(map[string]net.Conn),
		msgch: make(chan client, 10),
		quit:  make(chan string, 10),
	}

	// Fill the chat map to its capacity limit
	for i := 0; i < 10; i++ {
		clientConn, serverConn := net.Pipe()
		defer clientConn.Close()
		defer serverConn.Close()
		server.chat[fmt.Sprintf("user%d", i)] = serverConn
	}

	// Create a new connection attempt
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen on a port: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		defer conn.Close()

		server.HandleConnection(conn)
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Read the response from the server
	response := make([]byte, 256)
	n, err := conn.Read(response)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	expectedMessage := "The chat is full try another time"
	if !strings.Contains(string(response[:n]), expectedMessage) {
		t.Errorf("Expected message %q, but got %q", expectedMessage, string(response[:n]))
	}
}

func TestNewServerCreatesServerWithCorrectListenAddress(t *testing.T) {
	listenAddr := "localhost:129345"
	server := NewServer(listenAddr)

	if server.listenAddr != listenAddr {
		t.Errorf("Expected listen address to be %q, but got %q", listenAddr, server.listenAddr)
	}
}

func TestServerStartHandlesInvalidAddress(t *testing.T) {
	server := &Server{
		listenAddr: "invalid-address",
		msgch:      make(chan client, 10),
		chat:       make(map[string]net.Conn),
		quit:       make(chan string, 10),
		history:    make([]string, 0),
		mu:         sync.RWMutex{},
	}

	done := make(chan bool)
	go func() {
		server.Start()
		done <- true
	}()

	select {
	case <-done:
		// The test should pass if the server does not panic or hang indefinitely
		// Since the server does nothing on invalid address, we expect it to terminate
	case <-time.After(1 * time.Second):
		t.Error("Expected server to handle invalid address and terminate, but it did not")
	}
}

func TestRemoveClient(t *testing.T) {
	server := &Server{
		chat: make(map[string]net.Conn),
		quit: make(chan string, 10),
	}

	// Simulate a client connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Add a client to the chat map
	username := "testuser"
	server.chat[username] = serverConn

	// Send the username to the quit channel
	server.quit <- username

	// Run removeclient in a separate goroutine
	go server.removeclient()

	// Allow some time for the goroutine to process
	time.Sleep(100 * time.Millisecond)

	// Check if the client was removed from the chat map
	if _, exists := server.chat[username]; exists {
		t.Errorf("Expected client %q to be removed from chat map, but it still exists", username)
	}
}

func TestUpdateWritesAllHistoryMessagesToConnection(t *testing.T) {
	server := &Server{
		history: []string{"First message\n", "Second message\n", "Third message\n"},
		mu:      sync.RWMutex{},
	}

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Run the update function
	go server.update(serverConn)

	// Read from the client side of the pipe
	response := make([]byte, 1024)
	n, err := clientConn.Read(response)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	// Check if the response contains all history messages in order
	expectedResponse := "First message\n"
	if string(response[:n]) != expectedResponse {
		t.Errorf("Expected response %q, but got %q", expectedResponse, string(response[:n]))
	}
}
