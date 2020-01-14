package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	fileExtensionJSON = ".json"
)

// TODO: Implements configuration file encryption
const encryptedFileEnding = ".enc"

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func ensureDir(dirname string, perm os.FileMode) {
	if !fileExists(dirname) {
		_ = os.MkdirAll(dirname, perm)
	}
}

func findFile(file string, filename string) string {
	if len(filename) == 0 {
		if len(file) == 0 {
			log.Fatal("the filename is required")
		}
		dir := configDir()
		ensureDir(dir, 0700)
		basePath := filepath.Join(dir, file)
		switch {
		case fileExists(basePath + encryptedFileEnding):
			return basePath + encryptedFileEnding
		case fileExists(basePath + encryptedFileEnding + tmpExtension):
			return basePath + encryptedFileEnding
		}
		return basePath
	}
	ensureDir(filepath.Dir(filename), 0700)
	return filename
}

func findConfigFile(filename string) string {
	return findFile(fmt.Sprintf("config.%s", fileExtensionJSON), filename)
}

const tmpExtension = ".000~"

func safeWrite(name string, data []byte, perm os.FileMode) error {
	tempName := name + tmpExtension
	err := ioutil.WriteFile(tempName, data, perm)
	if err != nil {
		return err
	}

	if fileExists(name) {
		os.Remove(name)
	}

	return os.Rename(tempName, name)
}

func readFileOrTemporaryBackup(name string) (data []byte, e error) {
	if fileExists(name) {
		data, e = ioutil.ReadFile(name)
		if len(data) == 0 && fileExists(name+tmpExtension) {
			data, e = ioutil.ReadFile(name + tmpExtension)
		}
		return
	}
	return ioutil.ReadFile(name + tmpExtension)
}

func configDir() string {
	return filepath.Join(SystemConfigDir(), "tonio")
}

// SystemConfigDir returns the application data directory, valid on both windows and posix systems
func SystemConfigDir() string {
	return XdgConfigHome()
}
