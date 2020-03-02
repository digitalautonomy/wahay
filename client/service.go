package client

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

var (
	// ErrNoClient is throwed when no available client has been initialized
	ErrNoClient = errors.New("error: no client to run")
	// ErrNoService is used when the Tor service can't be started
	ErrNoService = errors.New("error: the service can't be started")
)

// LaunchClient executes the current Mumble client instance
func LaunchClient(data hosting.MeetingData, onClose func()) (tor.Service, error) {
	c := GetMumbleInstance()

	if !c.CanBeUsed() {
		return nil, ErrNoClient
	}

	err := c.LoadCertificateFrom(
		data.MeetingID,
		data.Port,
		data.Cert,
		hosting.DefaultCertificateServerPort)
	if err != nil {
		log.WithFields(log.Fields{
			"serviceID": data.MeetingID,
			"port":      hosting.DefaultCertificateServerPort,
		}).Errorf("No valid Mumble certificate available: %s", err)
		return nil, err
	}

	cm := tor.Command{
		Cmd:      c.GetBinaryPath(),
		Args:     []string{data.GenerateURL()},
		Modifier: c.GetTorCommandModifier(),
	}

	s, err := tor.NewService(cm)
	if err != nil {
		return nil, ErrNoService
	}

	s.OnClose(func() {
		c.Cleanup()

		if onClose != nil {
			onClose()
		}
	})

	if onClose != nil {
		s.OnClose(onClose)
	}

	return s, nil
}
