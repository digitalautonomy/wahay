package client

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"autonomia.digital/tonio/app/config"
)

// Instance is a representation of the Mumble client for Tonio
type Instance interface {
	GetBinary() string
	CanBeUsed() bool
	GetLastError() error
	// TODO[OB]: this is exposing too much information. It would be better that both
	//  tor and mumble have some kind of decorators for managing the commands
	GetBinaryEnv() []string
}

type client struct {
	sync.Mutex
	binary     string
	binaryPath string
	env        []string
	isValid    bool
	lastError  error
	configFile string
}

// TODO[OB]: this is very strange to me.
func (c *client) GetBinary() string {
	env := c.GetBinaryEnv()

	if len(env) == 0 {
		return c.binary
	}

	return strings.Join(env, c.binary)
}

func (c *client) CanBeUsed() bool {
	return c.isValid
}

func (c *client) GetLastError() error {
	return c.lastError
}

func (c *client) GetBinaryEnv() []string {
	return c.env
}

const (
	configFileName = "mumble.ini"
	configDataName = ".mumble.sqlite"
)

// InitSystem do the checking of the current system looking
// for the  appropriate Mumble binary and check for errors
func InitSystem(conf *config.ApplicationConfig) (Instance, error) {
	dirs := []string{
		conf.GetMumbleBinaryPath(),
		getDestinationFile(),
		// TODO[OB]: why hard code it here? What if it's on the $PATH somewhere else in the system???
		"/usr/bin/mumble",
	}

	// TODO[OB]: this puts the bundled mumble at the end of the possibilities. Not what we decided
	localDir, err := os.Getwd()
	if err == nil {
		dirs = append(dirs, filepath.Join(localDir, "mumble/mumble"))
	}

	binary, env := findMumbleBinary(dirs)

	if len(binary) == 0 {
		return nil, errors.New("client not found")
	}

	c := &client{
		binary:  binary,
		env:     env,
		isValid: true,
	}

	// If any valid binary is found, then we need to copy it
	// to a specific directory so we can use it with a custom
	// configuration for that Mumble client
	err = c.prepareToUseInOurDir()
	if err != nil {
		log.Printf("Client error: %s\n", err.Error())
		c.isValid = false
		c.lastError = err
	} else {
		c.binary = getDestinationFile()
		c.binaryPath = getDestinationDir()
		c.lastError = c.ensureDirectory()
	}

	return c, c.lastError
}

var (
	errClientBinaryNotFound    = errors.New("client binary not found")
	errClientBinaryInvalidCopy = errors.New("client invalid copying")
)

func getDestinationDir() string {
	dir := filepath.Join(config.Dir(), "client")
	config.EnsureDir(dir, 0700)
	return dir
}

func getDestinationFile() string {
	return filepath.Join(getDestinationDir(), "mumble")
}

func (c *client) prepareToUseInOurDir() error {
	c.Lock()
	defer c.Unlock()

	if !config.FileExists(c.binary) {
		return errClientBinaryNotFound
	}

	if c.binary == getDestinationFile() {
		return nil
	}

	err := copyBinToDir(c.binary, getDestinationFile())
	if err != nil {
		return errClientBinaryInvalidCopy
	}

	return nil
}

func copyBinToDir(source, destination string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(source); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(destination); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}

	if srcinfo, err = os.Stat(source); err != nil {
		return err
	}

	return os.Chmod(destination, srcinfo.Mode())
}

// TODO[OB]: we should probably put this outside, using ext to manage it, like the css and other things
const mumbleInitConfig = `
[General]
lastupdate=2

[net]
tcponly=true

[overlay]
enable=false
version=1.3.0

[privacy]
hideos=true

[shortcuts]
size=0

[ui]
WindowLayout=1
alwaysontop=1
askonquit=false
drag=1
language=es
usage=false
`

func (c *client) ensureDirectory() error {
	config.EnsureDir(c.binaryPath, 0700)
	config.EnsureDir(filepath.Join(c.binaryPath, "Overlay"), 0700)
	config.EnsureDir(filepath.Join(c.binaryPath, "Plugins"), 0700)
	config.EnsureDir(filepath.Join(c.binaryPath, "Themes"), 0700)

	if !config.FileExists(filepath.Join(c.binaryPath, configDataName)) {
		f, err := os.Create(filepath.Join(c.binaryPath, configDataName))
		if err != nil {
			return err
		}
		f.Close()
	}

	err := c.writeConfigToFile()
	if err != nil {
		return err
	}

	return nil
}

func (c *client) writeConfigToFile() error {
	if len(c.configFile) == 0 {
		c.configFile = filepath.Join(c.binaryPath, configFileName)
	}

	if config.FileExists(c.configFile) {
		return nil
	}

	c.Lock()
	defer c.Unlock()

	return config.SafeWrite(c.configFile, []byte(mumbleInitConfig), 0600)
}
