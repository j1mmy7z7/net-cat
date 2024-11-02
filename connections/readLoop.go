package netcat

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	Username string
	message  string
}
type Server struct {
	listenAddr string
	ln         net.Listener
	Msgch      chan Client
	Quit       chan string
	chat       map[string]net.Conn
	history    []string
	mu         sync.RWMutex
}

func (s *Server) readLoop(conn net.Conn, user *Client) {
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
