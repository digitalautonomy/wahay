package hosting

import (
	"errors"
	"net"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/tor"
)

// CreateServerCollection creates the hosting server
func CreateServerCollection() (Servers, error) {
	return create()
}

const (
	// DefaultPort is a representation of the default port Mumble server
	DefaultPort = 64738
)

// Based on Whonix best practices:
// http://www.dds6qkxpwdeubwucdiaord2xgbbeyds25rbsgr73tbfpqpt4a6vjwsyd.onion
// /wiki/Dev/Whonix_friendly_applications_best_practices#Listen_Interface
func defaultHost() string {
	allInterfaces := "0.0.0.0"
	localhostInterface := "127.0.0.1"

	// Based on https://stackoverflow.com/a/12518877
	switch _, err := os.Stat("/usr/share/anon-ws-base-files/workstation"); {
	case err == nil:
		// We're in a Whonix-like environment; listen on all interfaces.
		return allInterfaces
	case os.IsNotExist(err):
		// We're not in Whonix; listen on localhost only.
		return localhostInterface
	default:
		// Some kind of error occurred; we don't know if we're on Whonix.  Fall
		// back to non-Whonix default, which should at least be safe.
		log.Errorf("defaultHost(): %s", err)
	}

	return localhostInterface
}

var errInvalidPort = errors.New("invalid port supplied")

// SuperUserData is an struct that represents the superuser data
// of a Grumble server
type SuperUserData struct {
	Username string
	Password string
}

// Service is a representation of our custom Mumble server
type Service interface {
	ID() string
	URL() string
	Port() int
	ServicePort() int
	SetWelcomeText(string)
	NewConferenceRoom(password string, u SuperUserData) error
	Close() error
}

type service struct {
	port        int
	mumblePort  int
	welcomeText string
	onion       tor.Onion
	room        *conferenceRoom
	httpServer  *webserver
	collection  Servers
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

func (s *service) SetWelcomeText(t string) {
	s.welcomeText = t
}

type conferenceRoom struct {
	server Server
}

func (s *service) NewConferenceRoom(password string, u SuperUserData) error {
	serv, err := s.collection.CreateServer(
		setDefaultOptions,
		setWelcomeText(s.welcomeText),
		setPort(strconv.Itoa(s.port)),
		setPassword(password),
		setSuperUser(u.Username, u.Password),
	)
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
func (s *servers) NewService(port string, t tor.Instance) (Service, error) {
	var onionPorts []tor.OnionPort

	httpServer, err := newCertificateServer(s.DataDir())
	if err != nil {
		return nil, err
	}

	onionPorts = append(onionPorts, tor.OnionPort{
		DestinationHost: defaultHost(),
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
		DestinationHost: defaultHost(),
		DestinationPort: serverPort,
		ServicePort:     p,
	})

	onion, err := t.NewOnionServiceWithMultiplePorts(onionPorts)
	if err != nil {
		return nil, err
	}

	ss := &service{
		port:       serverPort,
		mumblePort: p,
		onion:      onion,
		httpServer: httpServer,
		collection: s,
	}

	return ss, nil
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

	s.collection.Cleanup()

	return nil
}
