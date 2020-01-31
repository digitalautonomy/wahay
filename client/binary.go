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

// TODO[OB] - no point in having this declared
const noMumbleBinary = ""

var errInvalidCommand = errors.New("invalid command")

// TODO[OB] - idiomatic Golang is to return another parameter, a boolean, if it can't find the binary
func findMumbleBinary(dirsToLook []string) (string, []string) {
	// TODO[OB] - no point in pre-declaring this
	var err error
	var env []string

	// TODO[OB] - idiomatic Golang is to NOT use the index but instead
	//   iterate over the values themselves
	for index := range dirsToLook {
		env, err = isAnAvailableMumbleBinaryIn(dirsToLook[index])
		if err == nil {
			return dirsToLook[index], env
		}
	}

	return noMumbleBinary, nil
}

// TODO[OB] - I don't like that this returns the environment to run it in
//   and there's no point in returning a specific error every time, instead just
//   return a boolean
//   it might be better to have an equivalent of the tor.Instance for the Mumble
//   that contains the optional lib directory info, etc
func isAnAvailableMumbleBinaryIn(binaryPath string) ([]string, error) {
	// TODO[OB]: this is not really the right way to check for it. Instead:
	// - check that the binary file exists
	// - check that it's executable
	// - try to execute it with -h
	// - DON'T check the output, just check if there's an error from running it or not

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
