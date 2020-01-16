package tor

import (
	"errors"
	"os/exec"
	"regexp"
)

// Connectivity is used to check whether Tor can connect in different ways
type Connectivity interface {
	Check() (total error, partial error)
}

type connectivity struct {
	checkBinary bool
	routePort   int
	controlPort int
}

// DefaultRoutePort is the port Tor uses by default
const DefaultRoutePort = 9050

// DefaultControlPort is the port Tor uses by default for the control port
const DefaultControlPort = 9051

// NewDefaultChecker will test whether the default ports can
// be reached and are appropriate for our use
func NewDefaultChecker() Connectivity {
	// This checks everything, including binaries against the default ports
	return NewChecker(true, DefaultRoutePort, DefaultControlPort)
}

// NewChecker can check connectivity on custom ports, and optionally
// avoid checking for binary compatibility
func NewChecker(checkBinary bool, routePort, controlPort int) Connectivity {
	return &connectivity{checkBinary, routePort, controlPort}
}

func (c *connectivity) checkTorBinaryExists() bool {
	cmd := exec.Command("tor", "--version")
	err := cmd.Run()
	return err == nil
}

func extractVersionFrom(s []byte) string {
	r := regexp.MustCompile(`Tor version (?P<Version>.*?)\.`)
	result := r.FindSubmatch(s)
	if len(result) != 2 {
		return ""
	}
	return string(result[1])
}

func (c *connectivity) checkTorBinaryCompatibility() bool {
	cmd := exec.Command("tor", "--version")
	if cmd.Run() != nil {
		return false
	}

	output, _ := cmd.Output()

	diff, err := compareVersions(extractVersionFrom(output), MinSupportedVersion)
	if err != nil {
		return false
	}

	return diff >= 0
}

// What are the steps:
// 1 - check if the Tor binary exists  TOR_BINARY_EXISTS
// 2  - check the version of the Tor binary (tor --version) TOR_BINARY_COMPATIBILITY
// 3 - try to connect to default Control Port  TOR_CONTROL_PORT_EXISTS
// 4   - try different forms of authentication  TOR_CONTROL_PORT_CAN_AUTH
// 5 - check the version again  TOR_CONTROL_PORT_COMPATIBILITY
// 6 - try to use default Tor to connect to check.torproject.org to   TOR_CAN_ROUTE
//   ensure we can route traffic over Tor

// if !TOR_BINARY_EXISTS: fail, we can't do anything. report to user
// if !TOR_BINARY_COMPATIBILITY: fail, we can't do anything. report to user
// if !TOR_CONTROL_PORT_EXISTS || !TOR_CONTROL_PORT_CAN_AUTH || !TOR_CONTROL_PORT_COMPATIBILITY || !TOR_CAN_ROUTE:
//    - USE OUR OWN INSTANCE
//    - re run tests 3-6, just to be on the safe
//       - if any of them fail, fail and report to user

// todo: we should also check that either Torify or Torsocks is installed

func (c *connectivity) Check() (total error, partial error) {
	if c.checkBinary {
		if !c.checkTorBinaryExists() {
			return errors.New("no Tor binary installed"), nil
		}
		if !c.checkTorBinaryCompatibility() {
			return errors.New("version of Tor installed too old"), nil
		}
	}
	// them we go on and check the Control port stuff

	return nil, nil
}
