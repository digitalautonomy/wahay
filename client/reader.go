package client

import (
	"os"

	"github.com/digitalautonomy/wahay/codegen"
)

// TODO[OB]: Lots of getters

func getDBFileContent() []byte {
	content := codegen.GetFileWithFallback(".mumble.sqlite", "client/files", FSString)
	return []byte(content)
}

func getIniFileContent() string {
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

// TODO[OB]: These functions do not do what their name says they do. That's not great.

func directoryExists(dir string) bool {
	return dirOrFileExists(dir)
}

func fileExists(filename string) bool {
	return dirOrFileExists(filename)
}

func dirOrFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	// TODO: file may or may not exist. We should see err for details.
	return false
}

func isADirectory(path string) bool {
	dir, err := os.Stat(path)
	if err != nil {
		return false
	}

	return dir.IsDir()
}

// TODO[OB]: Are there not other things than files or directories in all the file systems on the planet?

func isAFile(filename string) bool {
	return !isADirectory(filename)
}
