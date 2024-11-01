package netcat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type client struct {
	Username string
	message  string
}
type Server struct {
	listenAddr string
	listener   net.Listener
	Msgch      chan client
	Quit       chan string
	chat       map[string]net.Conn
	history    []string
	mu         sync.RWMutex
}

// NewServer creates a new instance of the Server struct.//+
// It initializes the server with the provided listen address and sets up necessary channels and data structures.
// listenAddr: The address on which the server will listen for incoming connections./
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		Msgch:      make(chan client, 10),
		chat:       make(map[string]net.Conn),
		Quit:       make(chan string, 10),
		history:    make([]string, 0),
		mu:         sync.RWMutex{},
	}
}

// Start begins listening for incoming TCP connections on the server's address.
// It initializes the listener and continuously accepts new connections.
// Each connection is handled concurrently in a separate goroutine.
func (server *Server) Start() {
	listener, err := net.Listen("tcp", server.listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	server.listener = listener

	go server.handlemessages()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go server.handleConnection(conn)
	}
}

// handleConnection manages a new client connection to the server.
// It performs initial setup, updates the client with chat history,
// and starts a read loop to handle incoming messages.
//
// conn: The net.Conn object representing the client's connection.
func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	if len(server.chat) < 10 {
		name, err := Welcome(conn, server)
		if err != nil {
			conn.Write([]byte("An error occured while setting you up"))
			return
		}
		server.update(conn)
		server.readLoop(conn, name)
		server.removeclient()
	} else {
		conn.Write([]byte("The chat is full try another time"))
		// conn.Close()
	}
}

func (server *Server) handlemessages() {
	for sender := range server.Msgch {
		server.broadcastMessage(sender)
		server.gethistory(sender.message)
	}
}

func (s *Server) removeclient() {
	for name := range s.Quit {
		delete(s.chat, name)
	}
}

func (s *Server) update(conn net.Conn) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, msg := range s.history {
		conn.Write([]byte(msg))
	}
}

func (server *Server) broadcastMessage(sender client) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	// Log the message before broadcasting
	logMessage(sender.message)

	for key := range server.chat {
		if key != sender.Username {
			server.chat[key].Write([]byte(sender.message))
		}
	}
}

func (server *Server) gethistory(msg string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.history = append(server.history, msg)
}

func Welcome(conn net.Conn, server *Server) (*client, error) {
	buf := make([]byte, 256)

	var penguin strings.Builder

	file, err := os.Open("penguin.txt")
	if err != nil {
		log.Fatal(err)

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		penguin.WriteString(scanner.Text() + "\n")
	}

	fmt.Fprintf(conn, penguin.String()+"\n")
	fmt.Fprintf(conn, "[ENTER YOUR NAME]: ")
	name, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	user := buf[:name]
	user = user[:len(user)-1]
	New_client := &client{Username: string(user), message: fmt.Sprintf("%s has joined the chat\n", user)}
	server.mu.Lock()
	server.chat[New_client.Username] = conn
	server.mu.Unlock()
	server.Msgch <- *New_client
	return New_client, nil
}

func (s *Server) readLoop(conn net.Conn, user *client) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			user.message = fmt.Sprintf("%s has left the chat\n", user.Username)
			s.Msgch <- *user
			s.Quit <- user.Username
			break
		}
		conn.Write([]byte("\x1b[1A"))
		conn.Write([]byte("\x1b[2K"))
		message := strings.TrimSpace(string(buf[:n]))
		if message == "" {
			continue
		}

		if message == "/Q" {
			user.message = fmt.Sprintf("%s has left the chat\n", user.Username)
			s.Msgch <- *user
			s.Quit <- user.Username
			break
		}

		now := time.Now()
		user.message = fmt.Sprintf("[%s][%s]:%s", now.Format("2006-01-02 15:04:05"), user.Username, string(buf[:n]))
		conn.Write([]byte(user.message))
		s.Msgch <- *user
	}
}
