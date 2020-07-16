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
	errInvalidConfigFileDir    = errors.New("invalid client configuration directory")
	errInvalidConfigFileDBFile = errors.New("invalid client data file")
	errInvalidConfigFile       = errors.New("invalid client configuration")

	mumbleFolders = []string{
		"Overlay",
		"Plugins",
		"Themes",
	}
)

func (c *client) pathToConfig() string {
	if len(c.configDir) == 0 {
		location := c.pathToBinary()
		if !isADirectory(location) {
			location = filepath.Dir(location)
		}
		c.configDir = location
	}
	return c.configDir
}

func (c *client) regenerateConfiguration() error {
	location := c.pathToConfig()

	err := os.Remove(filepath.Join(location, configFileName))
	if err != nil {
		log.Errorf("Mumble client regenerateConfiguration(): %s", err.Error())
	}

	err = os.Remove(filepath.Join(location, configDBName))
	if err != nil {
		log.Errorf("Mumble client regenerateConfiguration(): %s", err.Error())
	}

	return c.ensureConfiguration()
}

func (c *client) ensureConfiguration() error {
	c.Lock()
	defer c.Unlock()

	var err error

	err = c.ensureConfigurationDir()
	if err != nil {
		return errInvalidConfigFileDir
	}

	err = c.ensureConfigurationDBFile()
	if err != nil {
		return errInvalidConfigFileDBFile
	}

	err = c.ensureConfigurationFile()
	if err != nil {
		return errInvalidConfigFile
	}

	return nil
}

func (c *client) ensureConfigurationDir() error {
	location := c.pathToConfig()

	err := createDir(location)
	if err != nil {
		log.Errorf("Error creating config directory: %s", location)
		return err
	}

	for _, dir := range mumbleFolders {
		err = createDir(filepath.Join(location, dir))
		if err != nil {
			log.Debugf("Error creating Mumble folder: %s", filepath.Join(location, dir))
			return err
		}
	}

	c.configDir = location

	return nil
}

func (c *client) ensureConfigurationDBFile() error {
	err := createFile(filepath.Join(c.configDir, configDBName))
	if err != nil {
		log.Debugf("The Mumble data file could not be created: %s", err)
		return err
	}

	configData := c.databaseProvider()
	err = ioutil.WriteFile(filepath.Join(c.configDir, configDBName), configData, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) ensureConfigurationFile() error {
	filename := filepath.Join(c.configDir, configFileName)

	err := createFile(filename)
	if err != nil {
		return errInvalidConfigFileDBFile
	}

	err = c.writeConfigToFile(filename)
	if err != nil {
		return errInvalidConfigFile
	}

	return nil
}

const (
	configFileName = "mumble.ini"
	configDBName   = ".mumble.sqlite"
)

func (c *client) writeConfigToFile(path string) error {
	if pathExists(c.configFile) {
		err := os.Remove(c.configFile)
		if err != nil {
			log.Debug(fmt.Sprintf("writeConfigToFile(): %s", err.Error()))
		}
	}

	var configFile string
	if isADirectory(path) {
		configFile = filepath.Join(path, configFileName)
	} else {
		configFile = filepath.Join(filepath.Dir(path), configFileName)
	}

	if !pathExists(configFile) || !isAFile(configFile) {
		err := createFile(configFile)
		if err != nil {
			return errInvalidConfigFileDBFile
		}
	}

	err := config.SafeWrite(configFile, []byte(c.configContentProvider()), 0600)
	if err != nil {
		return err
	}

	c.configFile = configFile

	return nil
}

func (c *client) saveCertificateConfigFile() error {
	if !pathExists(c.configFile) {
		return errors.New("invalid mumble.ini file")
	}

	content, err := ioutil.ReadFile(c.configFile)
	if err != nil {
		return err
	}

	tmc, err := generateTemporaryMumbleCertificate()
	if err != nil {
		log.Debugf("Error generating temporary mumble certificate: %v, assigning empty string", err)
		tmc = ""
	}

	certSectionProp := strings.Replace(
		string(content),
		"#CERTIFICATE",
		fmt.Sprintf("certificate=%s", tmc),
		1,
	)

	langSection := strings.Replace(
		certSectionProp,
		"#LANGUAGE",
		fmt.Sprintf("language=%s", config.DetectLanguage().String()),
		1,
	)

	err = ioutil.WriteFile(c.configFile, []byte(langSection), 0600)
	if err != nil {
		return err
	}

	return nil
}
