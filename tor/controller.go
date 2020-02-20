package tor

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/digitalautonomy/wahay/config"
	"github.com/wybiral/torgo"
)

// Control is the interface for controlling the Tor instance on this system
type Control interface {
	SetPassword(string)
	UseCookieAuth()
	CreateNewOnionServiceWithMultiplePorts(ports []OnionPort) (serviceID string, err error)
	CreateNewOnionService(destinationHost string, destinationPort int, port int) (serviceID string, err error)
	DeleteOnionService(serviceID string) error
	DeleteOnionServices()
}

type controller struct {
	torHost  string
	torPort  int
	authType *authenticationMethod
	password string
	c        torgoController
	tc       func(string) (torgoController, error)
}

var onions = []string{}

// CreateController takes the Tor information given
// and returns a controlling interface
func CreateController(torHost string, torPort int) Control {
	f := func(v string) (torgoController, error) {
		return torgo.NewController(v)
	}

	var a authenticationMethod = authenticateNone

	return &controller{
		torHost:  torHost,
		torPort:  torPort,
		authType: &a,
		tc:       f,
		c:        nil,
	}
}

func (cntrl *controller) SetPassword(p string) {
	cntrl.password = p
	if len(p) > 0 {
		var a authenticationMethod = authenticatePassword(p)
		cntrl.authType = &a
	}
}

func (cntrl *controller) UseCookieAuth() {
	var a authenticationMethod = authenticateCookie
	cntrl.authType = &a
}

// OnionPort is a representation of the information to create a hidde
// service with support for multiple destination ports
type OnionPort struct {
	ServicePort     int
	DestinationPort int
	DestinationHost string
}

func (cntrl *controller) CreateNewOnionServiceWithMultiplePorts(ports []OnionPort) (serviceID string, err error) {
	tc, err := cntrl.getTorController()
	if err != nil {
		return
	}

	if cntrl.authType != nil {
		err = (*cntrl.authType)(tc)
		if err != nil {
			return
		}
	}

	invalidPorts := []string{}
	finalPorts := make(map[int]string)
	for _, p := range ports {
		host := net.JoinHostPort(p.DestinationHost, strconv.Itoa(p.DestinationPort))

		if !config.CheckPort(p.ServicePort) || !config.CheckPort(p.DestinationPort) {
			invalidPorts = append(invalidPorts, host)
			continue
		}

		finalPorts[p.ServicePort] = host
	}

	if len(finalPorts) == 0 {
		return "", errors.New("invalid source port")
	} else if len(invalidPorts) > 0 {
		return "", fmt.Errorf("some ports are invalid: %v", invalidPorts)
	}

	onion := &torgo.Onion{
		Ports:          finalPorts,
		PrivateKeyType: "NEW",
		PrivateKey:     "ED25519-V3",
	}

	err = tc.AddOnion(onion)
	if err != nil {
		return "", err
	}

	serviceID = fmt.Sprintf("%s.onion", onion.ServiceID)
	onions = append(onions, serviceID)

	return serviceID, nil
}

func (cntrl *controller) CreateNewOnionService(destinationHost string, destinationPort int,
	servicePort int) (serviceID string, err error) {
	p := OnionPort{
		ServicePort:     servicePort,
		DestinationPort: destinationPort,
		DestinationHost: destinationHost,
	}
	return cntrl.CreateNewOnionServiceWithMultiplePorts([]OnionPort{p})
}

func (cntrl *controller) DeleteOnionService(serviceID string) error {
	s := strings.TrimSuffix(serviceID, ".onion")
	err := cntrl.c.DeleteOnion(s)
	if err != nil {
		return err
	}

	for i := range onions {
		if onions[i] == serviceID {
			onions[i] = onions[len(onions)-1]
			onions[len(onions)-1] = ""
			onions = onions[:len(onions)-1]
			break
		}
	}

	return nil
}

func (cntrl *controller) DeleteOnionServices() {
	if len(onions) > 0 {
		for i := range onions {
			_ = cntrl.DeleteOnionService(onions[i])
		}
	}
	onions = []string{}
}

func (cntrl *controller) getTorController() (torgoController, error) {
	if cntrl.c != nil {
		return cntrl.c, nil
	}

	c, err := cntrl.tc(net.JoinHostPort(cntrl.torHost, strconv.Itoa(cntrl.torPort)))
	if err != nil {
		return nil, err
	}

	cntrl.c = c

	return c, nil
}
