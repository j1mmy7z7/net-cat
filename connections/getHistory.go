package netcat

func (s *Server) gethistory(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, msg)
}
