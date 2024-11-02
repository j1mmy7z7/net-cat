package netcat

func (s *Server) broadcastMessage(sender Client) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Log the message before broadcasting
	logMessage(sender.message)

	for key := range s.chat {
		if key != sender.Username {
			s.chat[key].Write([]byte(sender.message))
		}
	}
}
