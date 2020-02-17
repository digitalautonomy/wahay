package client

import (
	"errors"
	"log"
	"os"
	"path/filepath"

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

func (c *client) EnsureConfiguration() error {
	c.Lock()
	defer c.Unlock()

	binaryDir := c.GetBinaryPath()
	if !isADirectory(binaryDir) {
		binaryDir = filepath.Dir(binaryDir)
	}

	err := createDir(binaryDir)
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
