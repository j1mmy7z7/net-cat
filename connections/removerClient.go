package netcat

func (s *Server) removeClient() {
	for name := range s.Quit {
		delete(s.chat, name)
	}
}


