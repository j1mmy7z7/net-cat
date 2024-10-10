package main

import (
	"fmt"
	"log"
	"net"
	"time"
	//"strings"
	"sync"
)

type Quit struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan Quit
	msgch      chan string
	quit       chan string
	chat map[string]net.Conn
	text  		[]string
	mu 			sync.Mutex
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan Quit),
		msgch:      make(chan string, 10),
		chat:		make(map[string]net.Conn),
		quit:       make(chan string, 10),
		text:       make([]string, 0),
		mu: 		sync.Mutex{},		
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.acceptLoop()

	<-s.quitch
	close(s.msgch)

	fmt.Println(s.listenAddr)
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		if len(s.chat) < 10 {
			go func(conn net.Conn) {
				name , err := Welcome(conn, s)
				if err != nil {
					conn.Write([]byte("An error occured while setting you up"))
					fmt.Println(err)
					return
				}
				s.mu.Lock() 
				collect(s)
				update(s, conn)
				s.mu.Unlock()
				s.readLoop(conn, name)
				remove(s)
			}(conn)
		} else {
			conn.Write([]byte("The chat is full try another time"))
			conn.Close()
		}
	}
}

func remove(s *Server) {
	for name :=  range s.quit {
		delete(s.chat, name)
	}

}

func collect(s *Server) {
	for msg := range s.msgch {
		s.text = append(s.text, msg)
	}
}

func update(s *Server,  conn net.Conn) {
	for _, value := range s.text {
		conn.Write([]byte(value))
	}
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
	fmt.Fprintf(conn, penguin + "\n")
	fmt.Fprintf(conn, "[ENTER YOUR NAME]: ")
	name, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	user := buf[:name]
	user = user[:len(user) -1]
	s.msgch <- fmt.Sprintf("%s has joined the chat\n", user)
	s.chat[string(user)] = conn
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

	go func() {
		for msg := range server.msgch {
			server.mu.Lock()
			for users := range server.chat {
				server.chat[users].Write([]byte(msg))
			}
			server.mu.Unlock()
		}
	}()
	log.Fatal(server.Start())
}
