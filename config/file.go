package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	fileExtensionJSON       = ".json"
	encrytptedFileExtension = ".axx"
	appConfigFile           = "config" + fileExtensionJSON
	appEncryptedConfigFile  = "config" + encrytptedFileExtension
)

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

var libDirs = []string{"/lib", "/lib64", "/lib/x86_64-linux-gnu", "/lib64/x86_64-linux-gnu"}
var libPrefixes = []string{"", "/usr", "/usr/local"}
var libSuffixes = []string{"", "/torsocks"}

func allLibDirs() []string {
	result := make([]string, 0)
	for _, l := range libDirs {
		for _, lp := range libPrefixes {
			for _, ls := range libSuffixes {
				result = append(result, filepath.Join(lp, l, ls))
			}
		}
	}
	return result
}

// FindByName returns a file by name if exists (libtorsocks.so)
func FindFileByName(fileName string) (string, error) {
	for _, ld := range allLibDirs() {
		fn := filepath.Join(ld, fileName)
		if FileExists(fn) {
			return fn, nil
		}
	}

	return "", fmt.Errorf("file not found %s", fileName)
}
