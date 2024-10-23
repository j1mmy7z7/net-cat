package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	//"strings"
	"sync"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	msgch      chan string
	quit       chan string
	chat       map[string]net.Conn
	history    []string
	mu         sync.RWMutex
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		msgch:      make(chan string, 10),
		chat:       make(map[string]net.Conn),
		quit:       make(chan string, 10),
		history:    make([]string, 0),
		mu:         sync.RWMutex{},
	}
}

func (s *Server) Start() {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return
	}
	defer ln.Close()
	s.ln = ln

	go s.handlemessages()

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	if len(s.chat) < 10 {
		name, err := Welcome(conn, s)
		if err != nil {
			conn.Write([]byte("An error occured while setting you up"))
			return
		}
		s.update(conn)
		s.readLoop(conn, name)
		s.removeclient()
	} else {
		conn.Write([]byte("The chat is full try another time"))
		conn.Close()
	}
}

func (s *Server) handlemessages() {
	for msg := range s.msgch {
		s.broadcastMessage(msg)
		s.gethistory(msg)
	}
}

func (s *Server) removeclient() {
	for name := range s.quit {
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

func (s *Server) broadcastMessage(msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Log the message before broadcasting
	logMessage(msg)

	for key := range s.chat {
		s.chat[key].Write([]byte(msg))
	}
}

func (s *Server) gethistory(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, msg)
}

func Welcome(conn net.Conn, s *Server) (string, error) {
	buf := make([]byte, 256)

	var penguin strings.Builder

	file, err := os.Open("penguin.txt")
	if err != nil {
		log.Println(err)
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
		return "", err
	}
	user := buf[:name]
	user = user[:len(user)-1]
	s.mu.Lock()
	s.chat[string(user)] = conn
	s.mu.Unlock()
	s.msgch <- fmt.Sprintf("%s has joined the chat\n", user)
	return string(user), nil
}

func (s *Server) readLoop(conn net.Conn, user string) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			s.msgch <- fmt.Sprintf("%s has left the chat\n", user)
			s.quit <- user
			break
		}

		message := strings.TrimSpace(string(buf[:n]))

		if message == "/q" {
			s.msgch <- fmt.Sprintf("%s has left the chat\n", user)
			s.quit <- user
			break
		}


		now := time.Now()
		s.msgch <- fmt.Sprintf("[%s][%s]:%s", now.Format("2006-01-02 15:04:05"), user, string(buf[:n]))
	}
}

func main() {
	defaultPort := "8989"

	port := defaultPort
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	if _, err := strconv.Atoi(port); err != nil {
		fmt.Printf("Invalid port number: %v", err)
	}


	log.Printf("Starting server on port %s...", port)
	server := NewServer(":" + port)
	server.Start()
}
