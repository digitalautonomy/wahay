package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

var (
	errInvalidConfigDirectory = errors.New("invalid client configuration directory")
	errInvalidDataFile        = errors.New("invalid client data file")
	errInvalidConfig          = errors.New("invalid client configuration")

	mumbleFolders = []string{
		"Overlay",
		"Plugins",
		"Themes",
	}
)

func (c *client) regenerateConfiguration() error {
	var err error

	binaryDir := c.GetBinaryPath()
	if !isADirectory(binaryDir) {
		binaryDir = filepath.Dir(binaryDir)
	}

	// Removes the configuration file (.ini)
	err = os.Remove(filepath.Join(binaryDir, configFileName))
	if err != nil {
		log.Errorf("Mumble client regenerateConfiguration(): %s", err.Error())
	}

	// Removes the Murmur sqlite database
	err = os.Remove(filepath.Join(binaryDir, configDataName))
	if err != nil {
		log.Errorf("Mumble client regenerateConfiguration(): %s", err.Error())
	}

	return c.ensureConfiguration()
}

func (c *client) ensureConfiguration() error {
	c.Lock()
	defer c.Unlock()

	var err error

	binaryDir := c.GetBinaryPath()
	if !isADirectory(binaryDir) {
		binaryDir = filepath.Dir(binaryDir)
	}

	err = createDir(binaryDir)
	if err != nil {
		return errInvalidConfigDirectory
	}

	for _, dir := range mumbleFolders {
		err = createDir(filepath.Join(binaryDir, dir))
		if err != nil {
			log.Printf("Error creating Mumble folder: %s", filepath.Join(binaryDir, dir))
		}
	}

	err = createFile(filepath.Join(binaryDir, configDataName))
	if err != nil {
		log.Println("The Mumble data file could not be created")
	}

	configData := c.databaseProvider()
	err = ioutil.WriteFile(filepath.Join(binaryDir, configDataName), configData, 0644)
	if err != nil {
		return err
	}

	filename := filepath.Join(binaryDir, configFileName)

	err = createFile(filename)
	if err != nil {
		return errInvalidDataFile
	}

	err = c.writeConfigToFile(filename)
	if err != nil {
		return errInvalidConfig
	}

	return nil
}

const (
	configFileName = "mumble.ini"
	configDataName = ".mumble.sqlite"
)

func (c *client) writeConfigToFile(path string) error {
	if len(c.configFile) > 0 && fileExists(c.configFile) {
		_ = os.Remove(c.configFile)
	}

	var configFile string
	if isADirectory(path) {
		configFile = filepath.Join(path, configFileName)
	} else {
		configFile = filepath.Join(filepath.Dir(path), configFileName)
	}

	if !isAFile(configFile) || !fileExists(configFile) {
		err := createFile(configFile)
		if err != nil {
			return errInvalidDataFile
		}
	}

	err := config.SafeWrite(configFile, []byte(c.configContentProvider()), 0600)
	if err != nil {
		return err
	}

	c.configFile = configFile

	return nil
}

// TODO: this function needs revision
func (c *client) saveCertificateConfigFile(cert string) error {
	if len(c.configFile) == 0 || !fileExists(c.configFile) {
		return errors.New("invalid mumble.ini file")
	}

	content, err := ioutil.ReadFile(c.configFile)
	if err != nil {
		return err
	}

	certSectionProp := strings.Replace(
		string(content),
		"#CERTIFICATE",
		fmt.Sprintf("certificate=%s", cert),
		1,
	)

	err = ioutil.WriteFile(c.configFile, []byte(certSectionProp), 0644)
	if err != nil {
		return err
	}

	return nil
}
