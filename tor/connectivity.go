package tor

import (
	"errors"
	"net"
	"strconv"

	"github.com/digitalautonomy/wahay/config"
	log "github.com/sirupsen/logrus"
)

// basicConnectivity is used to check whether Tor can connect in different ways
type basicConnectivity interface {
	Check() (errTotal error, errPartial error)
}

type connectivity struct {
	host        string
	routePort   int
	controlPort int
	password    string
}

func newCustomChecker(host string, routePort, controlPort int) basicConnectivity {
	return newChecker(host, routePort, controlPort, "")
}

func newDefaultChecker() basicConnectivity {
	return newChecker(defaultControlHost, defaultSocksPort, defaultControlPort, *config.TorControlPassword)
}

// newChecker can check connectivity on custom ports, and optionally
// avoid checking for binary compatibility
func newChecker(host string, routePort, controlPort int, password string) basicConnectivity {
	return &connectivity{
		host:        host,
		routePort:   routePort,
		controlPort: controlPort,
		password:    password,
	}
}

func (c *connectivity) checkTorControlPortExists() bool {
	_, err := torgof.NewController(net.JoinHostPort(c.host, strconv.Itoa(c.controlPort)))
	return err == nil
}

func withNewTorgoController(where string, a authenticationMethod) authenticationMethod {
	return func(torgoController) error {
		tc, err := torgof.NewController(where)
		if err != nil {
			return err
		}
		return a(tc)
	}
}

func debuggingAuth(name string, a authenticationMethod) authenticationMethod {
	return func(tc torgoController) error {
		log.Debugf("Running auth test(%s)", name)
		res := a(tc)
		log.Debugf(" -- result of running auth test (%s) - %v", name, res)
		return res
	}
}
func (c *connectivity) checkTorControlAuth() bool {
	where := net.JoinHostPort(c.host, strconv.Itoa(c.controlPort))

	authCallback := authenticateAny(
		withNewTorgoController(where, debuggingAuth("none", authenticateNone)),
		withNewTorgoController(where, debuggingAuth("cookie", authenticateCookie)),
		withNewTorgoController(where, debuggingAuth("password", authenticatePassword(c.password))))

	return authCallback(nil) == nil
}

type checkTorResult struct {
	IsTor bool
	IP    string
}

func (c *connectivity) checkConnectionOverTor() bool {
	return httpf.CheckConnectionOverTor(c.host, c.routePort)
}

var (
	// ErrPartialTorNoControlPort is an error to be trown when a valid Tor
	// control port cannot be found
	ErrPartialTorNoControlPort = errors.New("no Tor control port found")

	// ErrPartialTorNoValidAuth is an error to be trown when the system
	// cannot authenticate to the Tor control port
	ErrPartialTorNoValidAuth = errors.New("no Tor control port valid authentication")

	// ErrFatalTorNoConnectionAllowed is a fatal error that it's trown when
	// the system cannot make a connection over the Tor network
	ErrFatalTorNoConnectionAllowed = errors.New("no connection over Tor allowed")
)

func (c *connectivity) Check() (errTotal error, errPartial error) {
	if !c.checkTorControlPortExists() {
		return nil, ErrPartialTorNoControlPort
	}

	if !c.checkTorControlAuth() {
		return nil, ErrPartialTorNoValidAuth
	}

	if !c.checkConnectionOverTor() {
		return ErrFatalTorNoConnectionAllowed, nil
	}

	return nil, nil
}
