package client

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const noMumbleBinary = ""

var errInvalidCommand = errors.New("invalid command")

func findMumbleBinary(dirsToLook []string) (string, []string) {
	var err error
	var env []string

	for index := range dirsToLook {
		env, err = isAnAvailableMumbleBinaryIn(dirsToLook[index])
		if err == nil {
			return dirsToLook[index], env
		}
	}

	return noMumbleBinary, nil
}

func isAnAvailableMumbleBinaryIn(binaryPath string) ([]string, error) {
	env := []string{}

	command := exec.Command(binaryPath, "-h")

	libDir := filepath.Join(filepath.Dir(binaryPath), "lib")
	if _, err := os.Stat(libDir); !os.IsNotExist(err) {
		command.Env = os.Environ()
		envLibPath := fmt.Sprintf("LD_LIBRARY_PATH=%s", libDir)
		command.Env = append(command.Env, envLibPath)
		env = append(env, envLibPath)
	}

	output, err := command.Output()
	if output == nil && err != nil {
		return nil, errInvalidCommand
	}

	re := regexp.MustCompile(`^Usage:\smumble\s\[options\]\s\[<url>\]`)
	found := re.FindString(strings.TrimSpace(string(output)))
	if len(found) == 0 {
		return nil, errInvalidCommand
	}

	return env, nil
}
