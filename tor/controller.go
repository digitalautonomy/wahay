package tor

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/wybiral/torgo"
)

// Control is the interface for controlling the Tor instance on this system
type Control interface {
	EnsureTorCompatibility() error
	CreateNewOnionService(destinationHost, destinationPort string, port string) (serviceID string, err error)
	DeleteOnionService(serviceID string) error
}

type controller struct {
	torHost  string
	torPort  string
	password string
	tc       func(string) (torgoController, error)
	c        torgoController
}

const MinSupportedVersion = "0.3.2"

func (cntrl *controller) EnsureTorCompatibility() error {
	tc, err := cntrl.tc(net.JoinHostPort(cntrl.torHost, cntrl.torPort))
	if err != nil {
		return err
	}

	err = tc.AuthenticatePassword(cntrl.password)
	if err != nil {
		return err
	}

	version, err := tc.GetVersion()
	if err != nil {
		return err
	}

	diff, err := compareVersions(version, MinSupportedVersion)
	if err != nil {
		return err
	}

	if diff < 0 {
		return errors.New("version of Tor is not compatible")
	}

	return err
}

func (cntrl *controller) DeleteOnionService(serviceID string) error {
	if serviceID == "" {
		return errors.New("the service ID cannot be empty")
	}

	s := strings.TrimSuffix(serviceID, ".onion")
	return cntrl.c.DeleteOnion(s)
}

func (cntrl *controller) CreateNewOnionService(destinationHost, destinationPort string,
	port string) (serviceID string, err error) {
	tc, err := cntrl.tc(net.JoinHostPort(cntrl.torHost, cntrl.torPort))

	if err != nil {
		return
	}

	cntrl.c = tc

	err = tc.AuthenticatePassword(cntrl.password)

	if err != nil {
		return
	}

	servicePort, err := strconv.ParseUint(port, 10, 16)

	if err != nil {
		err = errors.New("invalid source port")
		return
	}

	onion := &torgo.Onion{
		Ports: map[int]string{
			int(servicePort): net.JoinHostPort(destinationHost, destinationPort),
		},
		PrivateKeyType: "NEW",
		PrivateKey:     "ED25519-V3",
	}

	err = tc.AddOnion(onion)

	if err != nil {
		return "", err
	}

	serviceID = fmt.Sprintf("%s.onion", onion.ServiceID)
	log.Printf("Service ID created: %s", serviceID)

	return serviceID, nil
}

// CreateController takes the Tor information given and returns a
// controlling interface
func CreateController(torHost, torPort, password string) Control {
	f := func(v string) (torgoController, error) {
		return torgo.NewController(v)
	}
	return &controller{
		torHost:  torHost,
		torPort:  torPort,
		password: password,
		tc:       f,
		c:        nil,
	}
}
