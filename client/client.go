package client

//go:generate esc -o gen_client_files.go -pkg client -ignore "Makefile" files

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/tor"
)

// TODO[OB]: Why does the GUI manage the certificate?
// TODO[OB]: Why does the GUI care about the binary path?

// Instance is a representation of the Mumble client for Wahay
type Instance interface {
	CanBeUsed() bool
	GetLastError() error
	Execute(args []string, onClose func()) (tor.Service, error)
	LoadCertificateFrom(serviceID string, servicePort int, cert []byte, webPort int) error
	GetTorCommandModifier() tor.ModifyCommand
	Destroy()
}

type client struct {
	sync.Mutex
	binary                *binary
	isValid               bool
	configFile            string
	configContentProvider mumbleIniProvider
	databaseProvider      databaseProvider
	err                   error
	torCommandModifier    tor.ModifyCommand
}

func newMumbleClient(p mumbleIniProvider, d databaseProvider) *client {
	c := &client{
		binary:                nil,
		isValid:               false,
		configContentProvider: p,
		databaseProvider:      d,
		err:                   nil,
	}

	return c
}

// TODO: implement a proper way to do this singleton
var currentInstance *client

type mumbleIniProvider func() string
type databaseProvider func() []byte

// Mumble returns the current Mumble instance
func Mumble() Instance {
	return currentInstance
}

// InitSystem do the checking of the current system looking
// for the  appropriate Mumble binary and check for errors
func InitSystem(conf *config.ApplicationConfig) Instance {
	var err error

	currentInstance = newMumbleClient(rederMumbleIniConfig, readerMumbleDB)

	b := searchBinary(conf)

	if b == nil {
		return invalidInstance(errors.New("a valid binary of Mumble is no available in your system"))
	}

	if b.shouldBeCopied {
		err = b.copyTo(getTemporaryDestinationForMumble())
		if err != nil {
			return invalidInstance(err)
		}
	}

	err = currentInstance.setBinary(b)
	if err != nil {
		return invalidInstance(err)
	}

	err = currentInstance.ensureConfiguration()
	if err != nil {
		return invalidInstance(err)
	}

	log.Infof("Using Mumble located at: %s\n", currentInstance.pathToBinary())
	log.Infof("Using Mumble environment variables: %s\n", currentInstance.binaryEnv())

	return currentInstance
}

func invalidInstance(err error) Instance {
	invalidInstance := &client{
		isValid: false,
		err:     err,
	}

	return invalidInstance
}

func (c *client) Execute(args []string, onClose func()) (tor.Service, error) {
	cm := tor.Command{
		Cmd:      c.pathToBinary(),
		Args:     args,
		Modifier: c.GetTorCommandModifier(),
	}

	s, err := tor.NewService(cm)
	if err != nil {
		return nil, errors.New("error: the service can't be started")
	}

	s.OnClose(func() {
		err := c.regenerateConfiguration()
		if err != nil {
			log.Errorf("Mumble client Destroy(): %s", err.Error())
		}

		if onClose != nil {
			onClose()
		}
	})

	return s, nil
}

var errInvalidBinary = errors.New("invalid client binary")

func (c *client) validate() error {
	c.isValid = false

	if c.binary == nil {
		c.err = errInvalidBinary
		return c.err
	}

	if !c.binary.isValid {
		c.err = errInvalidBinary
		return c.err
	}

	c.err = nil
	c.isValid = true

	return nil
}

func (c *client) CanBeUsed() bool {
	return c.isValid && c.err == nil
}

func (c *client) pathToBinary() string {
	if c.isValid && c.binary != nil {
		return c.binary.path
	}
	return ""
}

func (c *client) binaryEnv() []string {
	if c.isValid && c.binary != nil {
		return c.binary.envIfBundle()
	}
	return nil
}

func (c *client) GetLastError() error {
	return c.err
}

func (c *client) setBinary(b *binary) error {
	if !b.isValid {
		return errors.New("the provided binary is not valid")
	}

	c.binary = b
	return c.validate()
}

func (c *client) GetTorCommandModifier() tor.ModifyCommand {
	if !c.CanBeUsed() {
		return nil
	}

	if c.torCommandModifier != nil {
		return c.torCommandModifier
	}

	env := c.binaryEnv()
	if len(env) == 0 {
		return nil
	}

	c.torCommandModifier = func(command *exec.Cmd) {
		command.Env = append(command.Env, env...)
	}

	return c.torCommandModifier
}

func (c *client) Destroy() {
	c.binary.destroy()
}

func getTemporaryDestinationForMumble() string {
	dir, err := ioutil.TempDir("", "mumble")
	if err != nil {
		return ""
	}

	return dir
}
