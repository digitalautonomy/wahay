package tor

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"autonomia.digital/tonio/app/config"
)

var PathTorBinary string

func Initialize(configPath string) string {
	return findTorBinary(configPath)
}

func findTorBinary(configPath string) string {
	pathTorFound := checkInConfiguredPath(configPath)
	if len(pathTorFound) > 0 {
		return pathTorFound
	}

	pathTorFound = checkInTonioDataDirectory()
	if len(pathTorFound) > 0 {
		return pathTorFound
	}

	pathCWD, err := os.Getwd()
	if err == nil {
		pathTorFound = checkInLocalDirectory(pathCWD)
		if len(pathTorFound) > 0 {
			return pathTorFound
		}

		pathTorFound = checkInExecutableDirectory(pathCWD)
		if len(pathTorFound) > 0 {
			return pathTorFound
		}
	}

	pathTorFound = checkInCurrentWorkingDirectory()
	if len(pathTorFound) > 0 {
		return pathTorFound
	}

	pathTorFound = checkInTonioBinary()
	if len(pathTorFound) > 0 {
		return pathTorFound
	}

	pathTorFound = checkInHomeExecutableDirectory()
	if len(pathTorFound) > 0 {
		return pathTorFound
	}

	pathTorFound = checkWithWhich()
	if len(pathTorFound) > 0 {
		return pathTorFound
	}

	return ""
}

func checkInConfiguredPath(configuredPath string) string {
	if isThereConfiguredTorBinary(configuredPath) {
		return configuredPath
	}
	return ""
}

func checkInTonioDataDirectory() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "tonio/tor")
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func checkInLocalDirectory(pathCWD string) string {
	pathToFind := filepath.Join(pathCWD, "/tor")
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func checkInExecutableDirectory(pathCWD string) string {
	pathToFind := filepath.Join(pathCWD, "/bin/tor")
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func checkInCurrentWorkingDirectory() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "/tor")
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func checkInTonioBinary() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "/bin/tonio/tor/tor")
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func checkInHomeExecutableDirectory() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "/bin/tonio/tor")
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func checkWithWhich() string {
	outputWhich, err := executeCmd("which", []string{"tor"})
	if outputWhich == nil || err != nil {
		return ""
	}

	pathToFind := strings.TrimSpace(string(outputWhich))
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return ""
}

func isThereConfiguredTorBinary(path string) bool {
	if path != "" {
		return checkTorVersionCompatibility(path)
	}
	return false
}

func executeCmd(path string, args []string) ([]byte, error) {
	cmd := exec.Command(path, args...)
	output, err := cmd.Output()
	if output == nil || err != nil {
		return nil, errors.New("invalid command")
	}
	return output, nil
}

func checkTorVersionCompatibility(path string) bool {
	output, err := executeCmd(path, []string{"--version"})
	if output == nil || err != nil {
		return false
	}

	diff, err := compareVersions(extractVersionFrom(output), MinSupportedVersion)
	if err != nil {
		return false
	}

	return diff >= 0
}

func extractVersionFrom(s []byte) string {
	r := regexp.MustCompile(`(\d+\.)(\d+\.)(\d+\.)(\d)`)
	result := r.FindStringSubmatch(string(s))

	if len(result) == 0 {
		return ""
	}

	return result[0]
}
