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
	GetTorController() (torgoController, error)
	EnsureTorCompatibility() (bool, bool, error)
	CreateNewOnionService(destinationHost, destinationPort string, port string) (serviceID string, err error)
	DeleteOnionService(serviceID string) error
	SetInstance(i *Instance)
	Close()
}

type controller struct {
	torHost  string
	torPort  string
	password string
	authType *authenticationMethod
	c        torgoController
	i        *Instance
	tc       func(string) (torgoController, error)
}

func (cntrl *controller) SetInstance(i *Instance) {
	cntrl.i = i
}

func (cntrl *controller) GetTorController() (torgoController, error) {
	if cntrl.c != nil {
		return cntrl.c, nil
	}

	c, err := cntrl.tc(net.JoinHostPort(cntrl.torHost, cntrl.torPort))
	if err != nil {
		return nil, err
	}

	cntrl.c = c

	return c, nil
}

func getAuthenticationMethod(tc torgoController, cntrl *controller) (authenticationMethod, error) {
	err := tc.AuthenticateNone()
	if err == nil {
		return authenticateNone, nil
	}

	err = tc.AuthenticateCookie()
	if err == nil {
		return authenticateCookie, nil
	}

	if len(cntrl.password) > 0 {
		err = tc.AuthenticatePassword(cntrl.password)
		if err == nil {
			return authenticatePassword(cntrl.password), nil
		}
	}

	addr := net.JoinHostPort(cntrl.torHost, cntrl.torPort)
	return authenticateNone, fmt.Errorf("cannot authenticate to the Tor Control Port on %s", addr)
}

func (cntrl *controller) EnsureTorCompatibility() (bool, bool, error) {
	tc, err := cntrl.GetTorController()
	if err != nil {
		return false, false, err
	}

	if cntrl.authType == nil {
		a, err := getAuthenticationMethod(tc, cntrl)
		if err != nil {
			log.Println(err)
		} else {
			cntrl.authType = &a
		}
	}

	err = (*cntrl.authType)(tc)
	if err != nil {
		return false, true, err
	}

	version, err := tc.GetVersion()
	if err != nil {
		return false, true, err
	}

	diff, err := compareVersions(version, MinSupportedVersion)
	if err != nil {
		return false, true, err
	}

	if diff < 0 {
		return false, false, errors.New("version of Tor is not compatible")
	}

	return false, true, nil
}

func (cntrl *controller) DeleteOnionService(serviceID string) error {
	s := strings.TrimSuffix(serviceID, ".onion")
	return cntrl.c.DeleteOnion(s)
}

func (cntrl *controller) CreateNewOnionService(destinationHost, destinationPort string,
	port string) (serviceID string, err error) {
	tc, err := cntrl.GetTorController()

	if err != nil {
		log.Println(err)
		return
	}

	err = (*cntrl.authType)(tc)
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

	return serviceID, nil
}

func (cntrl *controller) Close() {
	if cntrl.i != nil {
		cntrl.i.Destroy()
	}
}

// CreateController takes the Tor information given
// and returns a controlling interface
func CreateController(torHost, torPort, password string) Control {
	f := func(v string) (torgoController, error) {
		return torgo.NewController(v)
	}

	var a authenticationMethod = authenticateNone
	// If password is provided, then our `authType` should
	// be `authenticatePassword` as the default value
	if len(password) > 0 {
		a = authenticatePassword(password)
	}

	return &controller{
		torHost:  torHost,
		torPort:  torPort,
		password: password,
		authType: &a,
		tc:       f,
		c:        nil,
		i:        nil,
	}
}
