package netcat

import "net"

// handleConnection manages a new Client connection to the server.
// It performs initial setup, updates the Client with chat history,
// and starts a read loop to handle incoming messages.
// conn: The net.Conn object representing the Client's connection.
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
		s.removeClient()
	} else {
		conn.Write([]byte("The chat is full try another time"))
		conn.Close()
	}
}
