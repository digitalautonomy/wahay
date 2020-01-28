package client

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

const noMumbleBinary = ""

var errInvalidCommand = errors.New("invalid command")

var commonDirs = []string{
	"/usr/bin/mumble",
}

func getMumbleBinary(primaryDirs []string) string {
	dirsToLook := append(primaryDirs, commonDirs...)

	var err error
	for index := range dirsToLook {
		err = isAnAvailableMumbleBinaryIn(dirsToLook[index])
		if err == nil {
			return dirsToLook[index]
		}
	}

	return noMumbleBinary
}

func isAnAvailableMumbleBinaryIn(path string) error {
	command := exec.Command(path, "-h")

	output, err := command.Output()
	if output == nil && err != nil {
		log.Println(err)
		return errInvalidCommand
	}

	re := regexp.MustCompile(`^Usage:\smumble\s\[options\]\s\[<url>\]`)
	found := re.FindString(strings.TrimSpace(string(output)))
	if len(found) == 0 {
		return errInvalidCommand
	}

	return nil
}
