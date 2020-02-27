package tor

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/wybiral/torgo"
	"golang.org/x/net/proxy"
)

/*const UsrBinPath = "/usr/bin/tor"
const UsrLocalBinPath = "/usr/local/bin/tor"*/

// Connectivity is used to check whether Tor can connect in different ways
type Connectivity interface {
	Check() (errTotal error, errPartial error)
}

type connectivity struct {
	host        string
	routePort   int
	controlPort int
	password    string
}

// NewCustomChecker returns a Tor checker for our custom Tor instance
func NewCustomChecker(host string, routePort, controlPort int) Connectivity {
	return NewChecker(host, routePort, controlPort, "")
}

// NewChecker can check connectivity on custom ports, and optionally
// avoid checking for binary compatibility
func NewChecker(host string, routePort, controlPort int, password string) Connectivity {
	return &connectivity{
		host:        host,
		routePort:   routePort,
		controlPort: controlPort,
		password:    password,
	}
}

func (c *connectivity) checkTorControlPortExists() bool {
	_, err := torgo.NewController(net.JoinHostPort(c.host, strconv.Itoa(c.controlPort)))
	return err == nil
}

func withNewTorgoController(where string, a authenticationMethod) authenticationMethod {
	return func(torgoController) error {
		tc, err := torgo.NewController(where)
		if err != nil {
			return err
		}
		return a(tc)
	}
}

func (c *connectivity) checkTorControlAuth() bool {
	where := net.JoinHostPort(c.host, strconv.Itoa(c.controlPort))

	authCallback := authenticateAny(
		withNewTorgoController(where, authenticateNone),
		withNewTorgoController(where, authenticateCookie),
		withNewTorgoController(where, authenticatePassword(c.password)))

	return authCallback(nil) == nil
}

type checkTorResult struct {
	IsTor bool
	IP    string
}

func (c *connectivity) checkConnectionOverTor() bool {
	proxyURL, err := url.Parse("socks5://" + net.JoinHostPort(c.host, strconv.Itoa(c.routePort)))
	if err != nil {
		return false
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return false
	}

	t := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: t}

	resp, err := client.Get("https://check.torproject.org/api/ip")
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	var v checkTorResult
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return false
	}

	return v.IsTor
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
