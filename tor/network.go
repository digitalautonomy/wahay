package tor

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

var (
	defaultTorHost  = "127.0.0.1"
	defaultTorPorts = []string{"9051", "9150"}
	timeout         = 30 * time.Second
)

// State informs the state of Tor
type State interface {
	Detect() bool
	Address() string
	Host() string
	Port() string
	IsConnectionOverTor(proxy.Dialer) bool
}

// Network is the default state manager for Tor
var Network State = &defaultTorManager{}

type defaultTorManager struct {
	addr     string
	detected bool

	torHost  string
	torPort  string
	torPorts []string
}

func (m *defaultTorManager) Detect() bool {
	torHost := m.torHost
	if len(torHost) == 0 {
		torHost = defaultTorHost
	}

	torPorts := m.torPorts
	if len(m.torPorts) == 0 {
		torPorts = defaultTorPorts
	}

	var h string
	var p string
	var found bool
	m.addr, h, p, found = detectTor(torHost, torPorts)
	m.torHost = h
	m.torPort = p
	m.detected = found
	return found
}

func (m *defaultTorManager) Address() string {
	if !m.detected {
		m.Detect()
	}

	return m.addr
}

func (m *defaultTorManager) Host() string {
	if !m.detected {
		m.Detect()
	}

	return m.torHost
}

func (m *defaultTorManager) Port() string {
	if !m.detected {
		m.Detect()
	}

	return m.torPort
}

func detectTor(host string, ports []string) (string, string, string, bool) {
	for _, port := range ports {
		addr := net.JoinHostPort(host, port)
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			continue
		}

		defer conn.Close()
		return addr, host, port, true
	}

	return "", "", "", false
}

// CheckTorResult represents the JSON result from a check tor request
type CheckTorResult struct {
	IsTor bool
	IP    string
}

// IsConnectionOverTor will make a connection to the check.torproject page to see if we're using Tor or not
func (*defaultTorManager) IsConnectionOverTor(d proxy.Dialer) bool {
	if d == nil {
		d = proxy.Direct
	}

	c := &http.Client{Transport: &http.Transport{Dial: d.Dial}}

	resp, err := c.Get("https://check.torproject.org/api/ip")
	if err != nil {
		log.Printf("Got error when trying to check tor: %v", err)
		return false
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Got error when trying to check tor: %v", err)
		return false
	}

	v := CheckTorResult{}
	err = json.Unmarshal(content, &v)
	if err != nil {
		log.Printf("Got error when trying to check tor: %v", err)
		return false
	}

	return v.IsTor
}
