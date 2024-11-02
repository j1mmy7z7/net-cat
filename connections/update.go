package netcat

import "net"

func (s *Server) update(conn net.Conn) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, msg := range s.history {
		conn.Write([]byte(msg))
	}
}
