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
	check() (authType string, errTotal error, errPartial error)
}

type connectivity struct {
	host        string
	routePort   int
	controlPort int
	password    string
	authType    string
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

func (c *connectivity) settingAuthType(tp string, a authenticationMethod) authenticationMethod {
	return func(tc torgoController) error {
		res := a(tc)
		if res == nil {
			c.authType = tp
		}
		return res
	}
}

func (c *connectivity) checkTorControlAuth() bool {
	where := net.JoinHostPort(c.host, strconv.Itoa(c.controlPort))

	authCallback := authenticateAny(
		withNewTorgoController(where, c.settingAuthType("none", authenticateNone)),
		withNewTorgoController(where, c.settingAuthType("cookie", authenticateCookie)),
		withNewTorgoController(where, c.settingAuthType("password", authenticatePassword(c.password))))

	return authCallback(nil) == nil
}

func (c *connectivity) tryAuthenticate(tc torgoController) error {
	switch c.authType {
	case "none":
		return authenticateNone(tc)
	case "password":
		return authenticatePassword(c.password)(tc)
	case "cookie":
		return authenticateCookie(tc)
	default:
		return errors.New("no valid authentication type")
	}
}

func (c *connectivity) checkControlPortVersion() bool {
	where := net.JoinHostPort(c.host, strconv.Itoa(c.controlPort))

	tc, err := torgof.NewController(where)
	if err != nil {
		log.Debugf("checkControlPortVersion() - can't connect to control port: %v", err)
		return false
	}
	err = c.tryAuthenticate(tc)
	if err != nil {
		log.Debugf("checkControlPortVersion() - can't authenticate: %v", err)
		return false
	}

	v, err := tc.GetVersion()
	if err != nil {
		log.Debugf("checkControlPortVersion() - can't get version: %v", err)
		return false
	}

	diff, err := compareVersions(v, minSupportedVersion)
	if err != nil {
		log.Debugf("checkControlPortVersion() - can't compare versions: %v", err)
		return false
	}

	return diff >= 0
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

	// ErrPartialTorTooOld is an error that shows that the control port is running
	// a version that is too old
	ErrPartialTorTooOld = errors.New("the Tor control port is running a too old version of Tor")

	// ErrFatalTorNoConnectionAllowed is a fatal error that it's trown when
	// the system cannot make a connection over the Tor network
	ErrFatalTorNoConnectionAllowed = errors.New("no connection over Tor allowed")
)

func (c *connectivity) check() (authType string, errTotal error, errPartial error) {
	if !c.checkTorControlPortExists() {
		log.Debugf(" - no control port exists")
		return "", nil, ErrPartialTorNoControlPort
	}

	if !c.checkTorControlAuth() {
		log.Debugf(" - no valid authentication for control port")
		return "", nil, ErrPartialTorNoValidAuth
	}

	if !c.checkControlPortVersion() {
		log.Debugf(" - no valid version of tor on control port")
		return "", nil, ErrPartialTorTooOld
	}

	// While this returns ErrFatalTorNoConnectionAllowed as a total error
	// the System Tor checking will ignore this and not try to stop the
	// process. Thus the distinction between total and partial is only really
	// relevant for custom instances.

	if !c.checkConnectionOverTor() {
		log.Debugf(" - no connection over tor to the internet possible")
		return "", ErrFatalTorNoConnectionAllowed, nil
	}

	return c.authType, nil, nil
}
