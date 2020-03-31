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
	// IsValid returns a boolean indicating if the found client is valid to be used.
	// In case it returns false, the client last error should be checked.
	IsValid() bool

	// LastError returns the last error registered during client initialization
	LastError() error

	// Launch runs the found client through the Tor proxy with the given Mumble URL.
	// Before running the client the system will make a request of the certificate to the origin
	// based on the given url.
	Launch(url string, onClose func()) (tor.Service, error)

	Destroy()
}

type client struct {
	sync.Mutex
	binary                *binary
	isValid               bool
	configFile            string
	configDir             string
	configContentProvider mumbleIniProvider
	databaseProvider      databaseProvider
	err                   error
	torCmdModifier        tor.ModifyCommand
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
		tempDir, err := tempFolder()
		if err != nil {
			return invalidInstance(err)
		}

		err = b.copyTo(tempDir)
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

func (c *client) Launch(url string, onClose func()) (tor.Service, error) {
	// First, we load the certificate from the remote server and if a
	// valid certificate is found then we execute the client through Tor
	err := c.requestCertificate(url)
	if err != nil {
		log.WithFields(log.Fields{"url": url}).Errorf("Launch() client: %s", err.Error())
	}

	return c.execute([]string{url}, onClose)
}

func (c *client) execute(args []string, onClose func()) (tor.Service, error) {
	cm := tor.Command{
		Cmd:      c.pathToBinary(),
		Args:     args,
		Modifier: c.torCommandModifier(),
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

func (c *client) IsValid() bool {
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

func (c *client) LastError() error {
	return c.err
}

func (c *client) setBinary(b *binary) error {
	if !b.isValid {
		return errors.New("the provided binary is not valid")
	}

	c.binary = b
	return c.validate()
}

func (c *client) torCommandModifier() tor.ModifyCommand {
	if !c.IsValid() {
		return nil
	}

	if c.torCmdModifier != nil {
		return c.torCmdModifier
	}

	env := c.binaryEnv()
	if len(env) == 0 {
		return nil
	}

	c.torCmdModifier = func(command *exec.Cmd) {
		command.Env = append(command.Env, env...)
	}

	return c.torCmdModifier
}

func (c *client) Destroy() {
	c.binary.destroy()
}

func tempFolder() (string, error) {
	dir, err := ioutil.TempDir("", "mumble")
	if err != nil {
		return "", err
	}

	return dir, nil
}
