package client

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"sync"

	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/tor"
)

// Instance is a representation of the Mumble client for Tonio
type Instance interface {
	Invalidate(err error)
	CanBeUsed() bool
	GetBinaryPath() string
	GetBinaryEnv() []string
	GetLastError() error
	SetBinary(Binary) error
	Validate() error
	EnsureConfiguration() error
	GetTorCommandModifier() tor.ModifyCommand
	Destroy()
}

type client struct {
	sync.Mutex
	binary             Binary
	isValid            bool
	configFile         string
	err                error
	torCommandModifier tor.ModifyCommand
}

func newMumbleClient() *client {
	c := &client{
		binary:  nil,
		isValid: false,
		err:     nil,
	}

	return c
}

// InitSystem do the checking of the current system looking
// for the  appropriate Mumble binary and check for errors
func InitSystem(conf *config.ApplicationConfig) (c Instance) {
	var err error

	c = newMumbleClient()
	binary := getMumbleBinary(conf.GetMumbleBinaryPath())

	if binary == nil {
		c.Invalidate(errors.New("a valid binary of Mumble is no available in your system"))
		return
	}

	if binary.ShouldBeCopied() {
		err = binary.CopyTo(getTemporaryDestinationForMumble())
		if err != nil {
			c.Invalidate(err)
			return
		}
	}

	err = c.SetBinary(binary)
	if err != nil {
		c.Invalidate(err)
		return
	}

	err = c.EnsureConfiguration()
	if err != nil {
		c.Invalidate(err)
		return
	}

	return c
}

func (c *client) Invalidate(err error) {
	c.isValid = false
	c.err = err
}

var errInvalidBinary = errors.New("invalid client binary")

func (c *client) Validate() error {
	c.isValid = false

	if c.binary == nil {
		c.err = errInvalidBinary
		return c.err
	}

	if !c.binary.IsValid() {
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
		return c.binary.GetPath()
	}
	return ""
}

func (c *client) GetBinaryEnv() []string {
	if c.isValid && c.binary != nil {
		return c.binary.GetEnv()
	}
	return nil
}

func (c *client) GetLastError() error {
	return c.err
}

func (c *client) SetBinary(b Binary) error {
	if !b.IsValid() {
		return errors.New("the provided binary is not valid")
	}

	c.binary = b
	return c.Validate()
}

func (c *client) GetTorCommandModifier() tor.ModifyCommand {
	if !c.CanBeUsed() {
		return nil
	}

	if c.torCommandModifier != nil {
		return c.torCommandModifier
	}

	env := c.GetBinaryEnv()
	if len(env) == 0 {
		return nil
	}

	c.torCommandModifier = func(command *exec.Cmd) {
		command.Env = append(command.Env, env...)
	}

	return c.torCommandModifier
}

func (c *client) Destroy() {
	if c.binary.ShouldBeRemoved() {
		c.binary.Remove()
	}
}

func getTemporaryDestinationForMumble() string {
	dir, err := ioutil.TempDir(config.Dir(), "mumble")
	if err != nil {
		return ""
	}

	return dir
}
