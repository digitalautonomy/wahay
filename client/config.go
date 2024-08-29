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

	err = os.Remove(filepath.Join(location, configFileJSON))
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
	err := c.createAndWriteConfigFiles()
	if err != nil {
		return errInvalidConfigFile
	}

	return nil
}

func (c *client) createAndWriteConfigFiles() error {
	var configFileNames = map[string]func() string{
		configFileName: c.configContentProvider,
		configFileJSON: c.configJSONProvider,
	}

	for fileName, template := range configFileNames {
		filePath := filepath.Join(c.configDir, fileName)

		if pathExists(filePath) {
			err := os.Remove(filePath)
			if err != nil {
				log.Debug(fmt.Sprintf("writeConfigToFile(): %s", err.Error()))
			}
		}

		err := createFile(filePath)
		if err != nil {
			return err
		}

		err = c.writeConfigToFile(fileName, filePath, template)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	configFileJSON = "mumble_settings.json"
	configFileName = "mumble.ini"
	configDBName   = ".mumble.sqlite"
)

func (c *client) writeConfigToFile(fileName string, path string, template func() string) error {
	var configFile string
	if isADirectory(path) {
		configFile = filepath.Join(path, fileName)
	} else {
		configFile = filepath.Join(filepath.Dir(path), fileName)
	}

	if !pathExists(configFile) || !isAFile(configFile) {
		err := createFile(configFile)
		if err != nil {
			return errInvalidConfigFileDBFile
		}
	}

	dataBaseLocation := strings.Replace(
		template(),
		"#DATABASE",
		filepath.ToSlash(filepathJoin(c.configDir, configDBName)),
		1,
	)

	shortcutPtt := strings.Replace(
		dataBaseLocation,
		"#SHORTCUTPTT",
		ctrlRight,
		1,
	)

	language := config.DetectLanguage().String()
	if isIniConfigFile(configFile) {
		language = fmt.Sprintf("language=%s", language)
	}

	langSection := strings.Replace(
		shortcutPtt,
		"#LANGUAGE",
		language,
		1,
	)

	err := config.SafeWrite(configFile, []byte(langSection), 0600)
	if err != nil {
		return err
	}

	c.configFiles[configFile] = struct{}{}

	return nil
}

func (c *client) saveCertificateConfigFile() error {
	tmc, err := generateTemporaryMumbleCertificate()
	if err != nil {
		log.Debugf("Error generating temporary mumble certificate: %v, assigning empty string", err)
		tmc = ""
	}

	for configFile, _ := range c.configFiles {
		if !pathExists(configFile) {
			return errors.New("invalid mumble config file")
		}

		certificate := fmt.Sprintf(tmc)

		if isIniConfigFile(configFile) {
			certificate = fmt.Sprintf("certificate=%s", tmc)
		}

		content, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}

		certSectionProp := strings.Replace(
			string(content),
			"#CERTIFICATE",
			certificate,
			1,
		)

		err = ioutil.WriteFile(configFile, []byte(certSectionProp), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func isIniConfigFile(configFile string) bool {
	return filepath.Ext(configFile) == ".ini"
}
