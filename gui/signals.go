package gui

const (
	// StatusChanged is a custom signal executed when
	// the application status change
	StatusChanged = "status-changed"
)

// AddSignal adds an event listener to the ApplicationStatus struct instance
func (s *ApplicationStatus) AddSignal(e string, ch chan string) {
	if s.signals == nil {
		s.signals = make(map[string][]chan string)
	}
	if _, ok := s.signals[e]; ok {
		s.signals[e] = append(s.signals[e], ch)
	} else {
		s.signals[e] = []chan string{ch}
	}
}

// RemoveSignal removes an event listener from the ApplicationStatus struct instance
func (s *ApplicationStatus) RemoveSignal(e string, ch chan string) {
	if _, ok := s.signals[e]; ok {
		for i := range s.signals[e] {
			if s.signals[e][i] == ch {
				s.signals[e] = append(s.signals[e][:i], s.signals[e][i+1:]...)
				break
			}
		}
	}
}

// Emit emits an event on the Dog struct instance
func (s *ApplicationStatus) Emit(e string, t string) {
	if _, ok := s.signals[e]; ok {
		for _, handler := range s.signals[e] {
			go func(handler chan string) {
				handler <- t
			}(handler)
		}
	}
}
