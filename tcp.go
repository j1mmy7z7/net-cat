package main

import (
	"fmt"
	"log"
	"net"
	"time"
	//"strings"
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
	chat map[string]net.Conn
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan Quit),
		msgch:      make(chan string, 10),
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
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		name := Welcome(conn, s)
		go s.readLoop(conn, name)
	}
}

func Welcome(conn net.Conn, s *Server) string{
	buf := make([]byte, 15)
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
	name, _ := conn.Read(buf)
	user := buf[:name]
	user = user[:len(user) -1]
	s.msgch <- fmt.Sprintf("%s has joined the chat\n", user)
	return string(user)
}

func (s *Server) readLoop(conn net.Conn, user string) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			s.msgch <- fmt.Sprintf("%s has left the chat\n", user)
			break
		}
		now := time.Now()
		s.msgch <- fmt.Sprintf("[%s][%s] %s", now.Format("2006-01-02 15:04:05"), user, string(buf[:n])) 
	}
}

func main() {
	server := NewServer(":8081")

	go func() {
		for msg := range server.msgch {
			fmt.Print(msg)
		}
	}()
	log.Fatal(server.Start())
}
