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

// Instance is a representation of the Mumble client for Wahay
type Instance interface {
	CanBeUsed() bool
	GetBinaryPath() string
	GetLastError() error
	EnsureConfiguration() error
	LoadCertificateFrom(serviceID string, servicePort int, cert []byte, webPort int) error
	GetTorCommandModifier() tor.ModifyCommand
	Log()
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

var currentInstance *client

type mumbleIniProvider func() string
type databaseProvider func() []byte

// GetMumbleInstance returns the current Mumble instance
func GetMumbleInstance() Instance {
	return currentInstance
}

// InitSystem do the checking of the current system looking
// for the  appropriate Mumble binary and check for errors
func InitSystem(conf *config.ApplicationConfig) Instance {
	var err error

	currentInstance = newMumbleClient(getIniFileContent, getDBFileContent)

	b := getMumbleBinary(conf)

	if b == nil {
		currentInstance.invalidate(errors.New("a valid binary of Mumble is no available in your system"))
		return currentInstance
	}

	if b.shouldBeCopied {
		err = b.copyTo(getTemporaryDestinationForMumble())
		if err != nil {
			currentInstance.invalidate(err)
			return currentInstance
		}
	}

	err = currentInstance.setBinary(b)
	if err != nil {
		currentInstance.invalidate(err)
		return currentInstance
	}

	err = currentInstance.EnsureConfiguration()
	if err != nil {
		currentInstance.invalidate(err)
	}

	return currentInstance
}

func (c *client) invalidate(err error) {
	c.isValid = false
	c.err = err
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

func (c *client) GetBinaryPath() string {
	if c.isValid && c.binary != nil {
		return c.binary.getPath()
	}
	return ""
}

func (c *client) getBinaryEnv() []string {
	if c.isValid && c.binary != nil {
		return c.binary.getEnv()
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

	env := c.getBinaryEnv()
	if len(env) == 0 {
		return nil
	}

	c.torCommandModifier = func(command *exec.Cmd) {
		command.Env = append(command.Env, env...)
	}

	return c.torCommandModifier
}

func (c *client) Log() {
	log.Infof("Using Mumble located at: %s\n", c.GetBinaryPath())
	log.Infof("Using Mumble environment variables: %s\n", c.getBinaryEnv())
}

func (c *client) Destroy() {
	c.binary.cleanup()
}

func getTemporaryDestinationForMumble() string {
	dir, err := ioutil.TempDir("", "mumble")
	if err != nil {
		return ""
	}

	return dir
}
