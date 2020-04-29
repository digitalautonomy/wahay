package tor

// Service is a representation of a service running through Tor
type Service interface {
	Close()
	IsClosed() bool
	OnClose(func())
}

type service struct {
	rc *RunningCommand

	onCloseFunctions []func()

	finished          bool
	finishedWithError error
	finishChannel     chan bool
}

// NewService creates a new Tor command service
func (i *instance) NewService(cmd string, args []string, modifier ModifyCommand) (Service, error) {
	rc, err := i.exec(cmd, args, modifier)
	if err != nil {
		return nil, err
	}

	s := &service{
		rc:                rc,
		onCloseFunctions:  nil,
		finished:          false,
		finishedWithError: nil,
		finishChannel:     make(chan bool),
	}

	s.listenToFinish()

	return s, nil
}

func (s *service) IsClosed() bool {
	return s.finished
}

func (s *service) Close() {
	s.rc.CancelFunc()
}

func (s *service) OnClose(f func()) {
	s.onCloseFunctions = append(s.onCloseFunctions, f)
}

func (s *service) listenToFinish() {
	s.closeWhenFinish()

	go func() {
		e := execf.WaitCommand(s.rc.Cmd)
		s.finished = true
		s.finishedWithError = e
		s.finishChannel <- true
	}()
}

func (s *service) closeWhenFinish() {
	go func() {
		<-s.finishChannel
		// TODO[OB] - Unnecessary if statement
		if len(s.onCloseFunctions) > 0 {
			for _, f := range s.onCloseFunctions {
				f()
			}
			// TODO[OB] - you can just set it to nil
			s.onCloseFunctions = s.onCloseFunctions[:0]
		}
	}()
}
