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

// FileExists check if a specific file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// EnsureDir creates a directory if not exists
func EnsureDir(dirname string, perm os.FileMode) {
	if !FileExists(dirname) {
		_ = os.MkdirAll(dirname, perm)
	}
}

// FindFile search for a specific file in the config directory
func FindFile(file string, filename string) string {
	if len(filename) == 0 {
		if len(file) == 0 {
			log.Fatal("the filename is required")
		}
		dir := Dir()
		EnsureDir(dir, 0700)
		basePath := filepath.Join(dir, file)
		switch {
		case FileExists(basePath + encryptedFileEnding):
			return basePath + encryptedFileEnding
		case FileExists(basePath + encryptedFileEnding + tmpExtension):
			return basePath + encryptedFileEnding
		}
		return basePath
	}
	EnsureDir(filepath.Dir(filename), 0700)
	return filename
}

func findConfigFile(filename string) string {
	return FindFile(fmt.Sprintf("config%s", fileExtensionJSON), filename)
}

const tmpExtension = ".000~"

// SafeWrite is a helper function to write content on specific file
func SafeWrite(name string, data []byte, perm os.FileMode) error {
	tempName := name + tmpExtension
	err := ioutil.WriteFile(tempName, data, perm)
	if err != nil {
		return err
	}

	if FileExists(name) {
		os.Remove(name)
	}

	return os.Rename(tempName, name)
}

// ReadFileOrTemporaryBackup tries to load a specific file
func ReadFileOrTemporaryBackup(name string) (data []byte, e error) {
	if FileExists(name) {
		data, e = ioutil.ReadFile(name)
		if len(data) == 0 && FileExists(name+tmpExtension) {
			data, e = ioutil.ReadFile(name + tmpExtension)
		}
		return
	}
	return ioutil.ReadFile(name + tmpExtension)
}

// Dir returns the default config directory for Tonio
func Dir() string {
	return filepath.Join(SystemConfigDir(), "tonio")
}

// SystemConfigDir returns the application data directory, valid on both windows and posix systems
func SystemConfigDir() string {
	return XdgConfigHome()
}
