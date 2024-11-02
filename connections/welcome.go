package netcat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func Welcome(conn net.Conn, s *Server) (*Client, error) {
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
		return nil, err
	}
	user := buf[:name]
	user = user[:len(user)-1]
	New_Client := &Client{Username: string(user), message: fmt.Sprintf("%s has joined the chat\n", user)}
	s.mu.Lock()
	s.chat[string(New_Client.Username)] = conn
	s.mu.Unlock()
	s.Msgch <- *New_Client
	return New_Client, nil
}
