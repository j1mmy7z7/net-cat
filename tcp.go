package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
}

type Message struct {
	from    string
	payload net.Conn
}

type Server struct {
	listenAddr string
	listen         net.Listener
	users     chan struct{}
	userMutex sync.Mutex
	msgch      chan Message
	prevMessages []string
	// peerMap map[net.Addr]
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		users:     make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.listen = ln
	fmt.Printf("Listening on port %s\n", s.listenAddr)
	go s.hanleMessage()
	s.acceptConnections()
	<-s.prevMessages
	
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listen.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		fmt.Println("New connection to server", conn.RemoteAddr())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}

		// conn.Write([]byte("Thanks for chatting!"))//everytime to the terminal of the user

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}
	}
}

func main() {
	server := NewServer(":8081")

	go func() {
		for msg := range server.msgch {
			fmt.Printf("Receive message from connection (%s):%s", msg.from, string(msg.payload))
		}
	}()
	log.Fatal(server.Start())
}

func (s *Server) getUserName(conn net.Conn) string {
	conn.Write([]byte("Enter your name:"))
	scanner := bufio.NewReader(conn)
	name, err := scanner.ReadString('\n')
	if err != nil || strings.TrimSpace(name) == "" {
		conn.Close()
		return ""
	}
	name = strings.TrimSpace(name)
	s.userMutex.Lock()
	s.users[conn] = Client{conn, name}
	s.userMutex.Unlock()

	s.notifyClients(fmt.Sprintf("%s has left the group", name), conn)
	return name
}

func (s *Server) previousMessages(conn net.Conn) {
	s.userMutex.Lock()
	for _, message := range s.prevMessages {
		conn.Write(([]byte(message + "\n")))
	}
	s.userMutex.Unlock()
}


func (s *Server) handleMess(){
	for message := range s.msgch {
		s.prevMessages = append(s.prevMessages, message.from)
		s.userMutex.Lock()
		formattedMessage, senderConn := message.from, message.payload
		for conn, client := range s.users {
			if conn !=  {
				client.conn.Write([]byte(formattedMessage + "\n"))
			}
		}
	}
}