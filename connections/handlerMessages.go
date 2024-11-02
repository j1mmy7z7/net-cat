package netcat

func (s *Server) handlemessages() {
	for sender := range s.Msgch {
		s.broadcastMessage(sender)
		s.gethistory(sender.message)
	}
}

