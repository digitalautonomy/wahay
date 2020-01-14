package tor

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"autonomia.digital/tonio/app/config"

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
const ControllerHost = "127.0.0.1"
const ControllerPort = "9951"

func (cntrl *controller) EnsureTorCompatibility() error {
	tc, err := cntrl.tc(net.JoinHostPort(cntrl.torHost, cntrl.torPort))
	if err != nil {
		return err
	}

	err = tc.AuthenticateCookie()
	if err != nil {
		log.Fatalf("TOR new instance can not be launched: %s", err)
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

	err = tc.AuthenticateCookie()
	if err != nil {
		log.Fatalf("TOR new instance can not be authenticated by cookie: %s", err)
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
func CreateController() Control {
	f := func(v string) (torgoController, error) {
		c, err := torgo.NewController(v)
		if err != nil {
			return nil, err
		}

		hd := config.XdgConfigHome()
		c.CookieFile = fmt.Sprintf("%s/tonio/tor/data/control_auth_cookie", hd)

		return c, nil
	}
	return &controller{
		torHost:  ControllerHost,
		torPort:  ControllerPort,
		password: "",
		tc:       f,
		c:        nil,
	}
}
