package netcat

import (
	"log"
	"net"
)

// Start begins listening for incoming TCP connections on the server's address.
// It initializes the listener and continuously accepts new connections.
// Each connection is handled concurrently in a separate goroutine.
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
