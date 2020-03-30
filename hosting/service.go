package hosting

import (
	"errors"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/tor"
)

var collection Servers

func ensureServerCollection() error {
	var err error

	collection, err = Create()
	if err != nil {
		return err
	}

	return nil
}

const (
	// DefaultPort is a representation of the default port Mumble server
	DefaultPort = 64738

	defaultHost = "127.0.0.1"
)

var errInvalidPort = errors.New("invalid port supplied")

// Service is a representation of our custom Mumble server
type Service interface {
	ID() string
	URL() string
	Port() int
	ServicePort() int
	NewConferenceRoom(password string) error
	Close() error
}

type service struct {
	port       int
	mumblePort int
	onion      tor.Onion
	room       *conferenceRoom
	httpServer *webserver
}

func (s *service) ID() string {
	return s.onion.ID()
}

func (s *service) URL() string {
	if s.ServicePort() != DefaultPort {
		return net.JoinHostPort(s.ID(), strconv.Itoa(s.ServicePort()))
	}
	return s.ID()
}

func (s *service) Port() int {
	return s.port
}

func (s *service) ServicePort() int {
	return s.mumblePort
}

type conferenceRoom struct {
	server Server
}

func (s *service) NewConferenceRoom(password string) error {
	serv, err := collection.CreateServer(strconv.Itoa(s.port), password)
	if err != nil {
		return err
	}

	err = serv.Start()
	if err != nil {
		return err
	}

	s.room = &conferenceRoom{
		server: serv,
	}

	// Start our certification http server
	s.httpServer.start(func(err error) {
		// TODO: We must inform the user about this error in a proper way
		log.Fatalf("Mumble certificate HTTP server: %v", err)
	})

	return nil
}

func (r *conferenceRoom) close() error {
	return r.server.Stop()
}

// NewService creates a new hosting service
func NewService(port string) (Service, error) {
	err := ensureServerCollection()
	if err != nil {
		return nil, err
	}

	var onionPorts []tor.OnionPort

	httpServer, err := newCertificateServer(collection.DataDir())
	if err != nil {
		return nil, err
	}

	onionPorts = append(onionPorts, tor.OnionPort{
		DestinationHost: defaultHost,
		DestinationPort: httpServer.port,
		ServicePort:     certServerPort,
	})

	p := DefaultPort
	if port != "" {
		p, err = strconv.Atoi(port)
		if err != nil {
			return nil, errInvalidPort
		}
	}

	serverPort := config.GetRandomPort()

	onionPorts = append(onionPorts, tor.OnionPort{
		DestinationHost: defaultHost,
		DestinationPort: serverPort,
		ServicePort:     p,
	})

	onion, err := tor.NewOnionServiceWithMultiplePorts(onionPorts)
	if err != nil {
		return nil, err
	}

	s := &service{
		port:       serverPort,
		mumblePort: p,
		onion:      onion,
		httpServer: httpServer,
	}

	return s, nil
}

var (
	// ErrServerNoClosed is an error to return when the server can't be stopped
	ErrServerNoClosed = errors.New("the current server can't be stopped")
	// ErrServerOnionDelete is an error to return when the hidden service can't be deleted
	ErrServerOnionDelete = errors.New("the hidden service can't be deleted")
)

func (s *service) Close() error {
	var err error

	if s.httpServer != nil {
		err = s.httpServer.stop()
		if err != nil {
			log.Errorf("hosting stop http server: Close(): %s", err)
		}
	}

	if s.room != nil {
		err = s.room.close()
		if err != nil {
			log.Errorf("hosting stop server: Close(): %s", err)
			return ErrServerNoClosed
		}
	}

	if s.onion != nil {
		err = s.onion.Delete()
		if err != nil {
			log.Errorf("hosting delete hidden service: Close(): %s", err)
			return ErrServerOnionDelete
		}
	}

	collection.Cleanup()

	return nil
}
