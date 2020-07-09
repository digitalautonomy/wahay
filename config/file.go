package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	fileExtensionJSON       = ".json"
	fileExtensionLOG        = ".log"
	fileExtensionBACKUP     = ".bak"
	encrytptedFileExtension = ".axx"
	appConfigFile           = "config" + fileExtensionJSON
	appEncryptedConfigFile  = "config" + encrytptedFileExtension
	appConfigFileBackup     = "config" + fileExtensionBACKUP
	appLogFile              = "application" + fileExtensionLOG
)

var wahayDataDir = filepath.Join(XdgDataHome(), "wahay")

// EnsureFilesAndDir ensure Wahay's required files and/or directories
func EnsureFilesAndDir() {
	_ = os.MkdirAll(wahayDataDir, 0700)
}

// CreateTempDir creates a temp dir inside Wahay's data dir
func CreateTempDir(dir string) string {
	EnsureFilesAndDir()
	d, _ := ioutil.TempDir(wahayDataDir, dir)
	return d
}

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

const tmpExtension = ".000~"

// SafeWrite is a helper function to write content on specific file
func SafeWrite(name string, data []byte, perm os.FileMode) error {
	tempName := name + tmpExtension
	err := ioutil.WriteFile(tempName, data, perm)
	if err != nil {
		return err
	}

	if FileExists(name) {
		_ = os.Remove(name)
	}

	return os.Rename(tempName, name)
}

// ReadFileOrTemporaryBackup tries to load a specific file
func ReadFileOrTemporaryBackup(name string) (data []byte, e error) {
	if FileExists(name) {
		data, e = ioutil.ReadFile(filepath.Clean(name))
		if len(data) == 0 && FileExists(name+tmpExtension) {
			data, e = ioutil.ReadFile(filepath.Clean(name + tmpExtension))
		}
		return
	}
	return ioutil.ReadFile(filepath.Clean(name + tmpExtension))
}

// Dir returns the default config directory for Wahay
func Dir() string {
	return filepath.Join(SystemConfigDir(), "wahay")
}

// TorDir returns the directory path for Tor
func TorDir() string {
	return filepath.Join(Dir(), "tor")
}

// SystemConfigDir returns the application data directory, valid on both windows and posix systems
func SystemConfigDir() string {
	return XdgConfigHome()
}

// RemoveAll removes a directory and it's children
func RemoveAll(dir string) error {
	return os.RemoveAll(dir)
}
