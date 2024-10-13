package main

import (
	"fmt"
	"log"
	"net"
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
	penguin := `Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    ` + "`" + `.       | ` + "`" + `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     ` + "`" + `-'       ` + "`" + `--'`
	fmt.Fprintf(conn, penguin+"\n")
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
		now := time.Now()
		s.msgch <- fmt.Sprintf("[%s][%s]:%s", now.Format("2006-01-02 15:04:05"), user, string(buf[:n]))
	}
}

func main() {
	server := NewServer(":8081")
	server.Start()
}
