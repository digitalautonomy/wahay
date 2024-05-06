package client

import (
	_ "embed"
	"os"
)

//go:embed files/.mumble.sqlite
var mumbleDBContent []byte

func readerMumbleDB() []byte {
	return mumbleDBContent
}

//go:embed files/mumble.ini
var mumbleIniContent string

func readerMumbleIniConfig() string {
	return mumbleIniContent
}

var osMkdirAll = os.MkdirAll

func createDir(path string) error {
	if pathExists(path) {
		return nil
	}

	return osMkdirAll(path, 0700)
}

var osCreate = os.Create

func createFile(filename string) error {
	if pathExists(filename) {
		return nil
	}

	file, err := osCreate(filename)
	if err != nil {
		return err
	}
	_ = file.Close()

	return nil
}

func pathExists(dir string) bool {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func isADirectory(path string) bool {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return false
	}

	return true
}

func isAFile(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
