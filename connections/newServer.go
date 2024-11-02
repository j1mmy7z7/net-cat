package netcat

import (
	"net"
	"sync"
)

// NewServer creates a new instance of the Server struct.//+
// It initializes the server with the provided listen address and sets up necessary channels and data structures.
// listenAddr: The address on which the server will listen for incoming connections./
func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		Msgch:      make(chan Client, 10),
		chat:       make(map[string]net.Conn),
		Quit:       make(chan string, 10),
		history:    make([]string, 0),
		mu:         sync.RWMutex{},
	}
}
