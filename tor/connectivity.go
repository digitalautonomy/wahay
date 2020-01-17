package tor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"

	"github.com/wybiral/torgo"
	"golang.org/x/net/proxy"
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
}

// DefaultHost is where Tor is hosted
const DefaultHost = "127.0.0.1"

// DefaultRoutePort is the port Tor uses by default
const DefaultRoutePort = 9050

// DefaultControlPort is the port Tor uses by default for the control port
const DefaultControlPort = 9051

// NewDefaultChecker will test whether the default ports can
// be reached and are appropriate for our use
func NewDefaultChecker() Connectivity {
	// This checks everything, including binaries against the default ports
	return NewChecker(true, DefaultHost, DefaultRoutePort, DefaultControlPort)
}

// NewChecker can check connectivity on custom ports, and optionally
// avoid checking for binary compatibility
func NewChecker(checkBinary bool, host string, routePort, controlPort int) Connectivity {
	return &connectivity{
		host:        host,
		checkBinary: checkBinary,
		routePort:   routePort,
		controlPort: controlPort,
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
	port := fmt.Sprintf("%d", c.controlPort)
	_, err := torgo.NewController(net.JoinHostPort(c.host, port))
	fmt.Println(err)
	return err == nil
}

func (c *connectivity) checkTorControlAuth() bool {
	port := fmt.Sprintf("%d", c.controlPort)
	tc, err := torgo.NewController(net.JoinHostPort(c.host, port))

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	authCallback := authenticateAny(authenticateNone, authenticateCookie, authenticatePassword(""))
	err = authCallback(tc)

	return err == nil
}

type checkTorResult struct {
	IsTor bool
	IP    string
}

func (c *connectivity) checkConnectionOverTor() bool {
	d, r := getTorProxy(c.host, c.routePort)
	if d == nil || !r {
		return false
	}

	cl := &http.Client{Transport: &http.Transport{Dial: d.Dial}}

	resp, err := cl.Get("https://check.torproject.org/api/ip")
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	v := checkTorResult{}
	err = json.Unmarshal(content, &v)
	if err != nil {
		return false
	}

	return v.IsTor
}

func getTorProxy(host string, port int) (proxy.Dialer, bool) {
	var dialer proxy.Dialer
	var err error

	u, e := url.Parse(fmt.Sprintf("socks5://%s:%d", host, port))
	if e != nil {
		return nil, true
	}

	if dialer, err = proxy.FromURL(u, dialer); err != nil {
		return nil, true
	}

	return dialer, false
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
