package gui

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"autonomia.digital/tonio/app/config"
)

func (u *gtkUI) initLogs() {
	if u.config == nil {
		return
	}

	if !u.config.IsLogsEnabled() {
		return
	}

	rawLogFile := ensureLogFile(u.config)
	if rawLogFile != nil {
		log.SetOutput(rawLogFile)
	}
}

func ensureLogFile(conf *config.ApplicationConfig) io.Writer {
	configuredLogFile := conf.GetRawLogFile()

	if len(configuredLogFile) == 0 {
		configuredLogFile = config.GetDefaultLogFile()
	}

	rawLogFile, err := ensureLogFileDirectory(configuredLogFile)
	if len(rawLogFile) == 0 || err != nil {
		return nil
	}

	file, err := os.OpenFile(rawLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}

	return file
}

func ensureLogFileDirectory(rawLogFile string) (string, error) {
	fileInfo, err := getLogFileInfo(rawLogFile, config.GetDefaultLogFile())
	if err != nil {
		return "", err
	}

	if fileInfo.IsDir() {
		return filepath.Join(rawLogFile, config.GetDefaultLogFileName()), nil
	}

	return rawLogFile, nil
}

func getLogFileInfo(rawLogFile string, defaultFile string) (os.FileInfo, error) {
	file, err := os.Stat(rawLogFile)

	if err != nil {
		if rawLogFile != defaultFile {
			file, err = os.Stat(defaultFile)
		}
	}

	return file, err
}
