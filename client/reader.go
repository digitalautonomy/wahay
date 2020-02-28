package client

import (
	"os"

	"github.com/digitalautonomy/wahay/codegen"
)

func getDBFileContent() []byte {
	content := codegen.GetFileWithFallback(".mumble.sqlite", "client/files", FSString)
	return []byte(content)
}

func getIniFileContent() string {
	return codegen.GetFileWithFallback("mumble.ini", "client/files", FSString)
}

func createDir(path string) error {
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

func isAFile(filename string) bool {
	return !isADirectory(filename)
}
