package tor

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"

	"autonomia.digital/tonio/app/config"
	"github.com/wybiral/torgo"
)

// Connectivity is used to check whether Tor can connect in different ways
type Connectivity interface {
	Check() (total error, partial error)
}

type connectivity struct {
	host        string
	checkBinary bool
	routePort   int
	controlPort int
	password    string
}

// NewDefaultChecker will test whether the default ports can
// be reached and are appropriate for our use
func NewDefaultChecker() Connectivity {
	// This checks everything, including binaries against the default ports
	return NewChecker(true, *config.TorHost, *config.TorRoutePort, *config.TorPort, *config.TorControlPassword)
}

// NewChecker can check connectivity on custom ports, and optionally
// avoid checking for binary compatibility
func NewChecker(checkBinary bool, host string, routePort, controlPort int, password string) Connectivity {
	return &connectivity{
		host:        host,
		checkBinary: checkBinary,
		routePort:   routePort,
		controlPort: controlPort,
		password:    password,
	}
}

func (c *connectivity) checkTorBinaryExists() bool {
	cmd := exec.Command("tor", "--version")
	err := cmd.Run()
	return err == nil
}

func extractVersionFrom(s []byte) string {
	r := regexp.MustCompile(`(\d+\.)(\d+\.)(\d+\.)(\d)`)
	result := r.FindStringSubmatch(string(s))

	if len(result) == 0 {
		return ""
	}

	return result[0]
}

func (c *connectivity) checkTorBinaryCompatibility() bool {
	return c.checkTorVersionCompatibility()
}

func (c *connectivity) checkTorVersionCompatibility() bool {
	cmd := exec.Command("tor", "--version")
	output, err := cmd.Output()
	if output == nil || err != nil {
		return false
	}

	diff, err := compareVersions(extractVersionFrom(output), MinSupportedVersion)
	if err != nil {
		return false
	}

	return diff >= 0
}

func (c *connectivity) checkTorControlPortExists() bool {
	_, err := torgo.NewController(net.JoinHostPort(c.host, strconv.Itoa(c.controlPort)))
	return err == nil
}

func (c *connectivity) checkTorControlAuth() bool {
	tc, err := torgo.NewController(net.JoinHostPort(c.host, strconv.Itoa(c.controlPort)))
	if err != nil {
		return false
	}

	authCallback := authenticateAny(authenticateNone, authenticateCookie, authenticatePassword(c.password))
	err = authCallback(tc)

	return err == nil
}

type checkTorResult struct {
	IsTor bool
	IP    string
}

func (c *connectivity) checkConnectionOverTor() bool {
	p := fmt.Sprintf("%d", c.routePort)
	cmd := exec.Command("torsocks", "curl", "https://check.torproject.org/api/ip", "-P", p)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	v := checkTorResult{}
	err = json.Unmarshal(output, &v)
	if err != nil {
		return false
	}

	return v.IsTor
}

func (c *connectivity) Check() (total error, partial error) {
	if c.checkBinary {
		if !c.checkTorBinaryExists() {
			return errors.New("no Tor binary installed"), nil
		}
		if !c.checkTorBinaryCompatibility() {
			return errors.New("version of Tor installed too old"), nil
		}
	}

	if !c.checkTorControlPortExists() {
		return nil, errors.New("no Tor Control Port found")
	}

	if !c.checkTorControlAuth() {
		return nil, errors.New("no Tor Control Port valid authentication")
	}

	if !c.checkBinary {
		if !c.checkTorVersionCompatibility() {
			return errors.New("version of Tor installed too old"), nil
		}
	}

	if !c.checkConnectionOverTor() {
		return errors.New("not connection over Tor allowed"), nil
	}

	return nil, nil
}
