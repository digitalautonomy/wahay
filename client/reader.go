package client

import (
	"os"

	"github.com/digitalautonomy/wahay/codegen"
)

func readerMumbleDB() []byte {
	content := codegen.GetFileWithFallback(".mumble.sqlite", "client/files", FSString)
	return []byte(content)
}

func rederMumbleIniConfig() string {
	return codegen.GetFileWithFallback("mumble.ini", "client/files", FSString)
}

func createDir(path string) error {
	if directoryExists(path) {
		return nil
	}

	return os.MkdirAll(path, 0700)
}

func createFile(filename string) error {
	if fileExists(filename) {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}

// TODO[OB]: sorry, but these functions STILL Doesn't do
// what their names say. The clue is that directoryExists and fileExists
// have the same implementation.

func directoryExists(dir string) bool {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func isADirectory(path string) bool {
	dir, err := os.Stat(path)
	if err != nil {
		return false
	}
	return dir.IsDir()
}

func isAFile(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}
