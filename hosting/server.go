package hosting

import grumbleServer "github.com/digitalautonomy/grumble/server"

// Server serves
type Server interface {
	Start() error
	Stop() error
}

type server struct {
	serverCollection *servers
	gs               *grumbleServer.Server
}

func (s *server) Start() error {
	err := s.gs.Start()
	if err != nil {
		return err
	}

	s.serverCollection.startListener()

	return nil
}

func (s *server) Stop() error {
	return s.gs.Stop()
}
