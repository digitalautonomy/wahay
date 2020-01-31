package tor

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"autonomia.digital/tonio/app/config"
)

const noPath = ""
const libTorsocks = "libtorsocks.so.0.0.0"

var libDirs = []string{"/lib", "/lib64", "/lib/x86_64-linux-gnu", "/lib64/x86_64-linux-gnu"}
var libPrefixes = []string{"", "/usr", "/usr/local"}
var libSuffixes = []string{"", "/torsocks"}

// Initialize find a Tor binary that can be used by Tonio
func Initialize(configPath string) (pathOfTorBinary string, foundInBundle bool) {
	return findTorBinary(configPath)
}

func findTorBinary(configPath string) (pathOfTorBinary string, foundInBundle bool) {
	pathTorFound := checkInConfiguredPath(configPath)
	if pathTorFound != noPath {
		return pathTorFound, false
	}

	pathTorFound = checkInTonioDataDirectory()
	if pathTorFound != noPath {
		return pathTorFound, false
	}

	pathCWD, err := os.Getwd()
	if err == nil {
		pathTorFound = checkInLocalDirectory(pathCWD)
		if pathTorFound != noPath {
			return pathTorFound, true
		}

		pathTorFound = checkInExecutableDirectory(pathCWD)
		if pathTorFound != noPath {
			return pathTorFound, true
		}

		pathTorFound = checkInExecutableDirectoryTor(pathCWD)
		if pathTorFound != noPath {
			return pathTorFound, true
		}
	}

	pathTorFound = checkInCurrentWorkingDirectory()
	if pathTorFound != noPath {
		return pathTorFound, false
	}

	pathTorFound = checkInTonioBinary()
	if pathTorFound != noPath {
		return pathTorFound, false
	}

	pathTorFound = checkInHomeExecutableDirectory()
	if pathTorFound != noPath {
		return pathTorFound, false
	}

	pathTorFound = checkWithWhich()
	if pathTorFound != noPath {
		return pathTorFound, false
	}

	return noPath, false
}

func checkInConfiguredPath(configuredPath string) string {
	log.Printf("checkInConfiguredPath... %s\n", configuredPath)
	if isThereConfiguredTorBinary(configuredPath) {
		return configuredPath
	}
	return noPath
}

func checkInTonioDataDirectory() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "tonio/tor")
	log.Printf("checkInTonioDataDirectory... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkInLocalDirectory(pathCWD string) string {
	pathToFind := filepath.Join(pathCWD, "/tor")
	log.Printf("checkInLocalDirectory... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkInExecutableDirectory(pathCWD string) string {
	pathToFind := filepath.Join(pathCWD, "/bin/tor")
	log.Printf("checkInExecutableDirectory... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkInCurrentWorkingDirectory() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "/tor")
	log.Printf("checkInCurrentWorkingDirectory... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkInExecutableDirectoryTor(pathCWD string) string {
	pathToFind := filepath.Join(pathCWD, "/tor/tor")
	log.Printf("checkInExecutableDirectoryTor... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkInTonioBinary() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "/bin/tonio/tor/tor")
	log.Printf("checkInTonioBinary... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkInHomeExecutableDirectory() string {
	pathToFind := filepath.Join(config.XdgDataHome(), "/bin/tonio/tor")
	log.Printf("checkInHomeExecutableDirectory... %s", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func checkWithWhich() string {
	outputWhich, err := executeCmd("which", []string{"tor"})
	if outputWhich == nil || err != nil {
		return noPath
	}

	pathToFind := strings.TrimSpace(string(outputWhich))
	log.Printf("checkWithWhich... %s\n", pathToFind)
	if isThereConfiguredTorBinary(pathToFind) {
		return pathToFind
	}
	return noPath
}

func isThereConfiguredTorBinary(path string) bool {
	if path != noPath {
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

// FindLibTorsocks returns the path of libtorsocks it exist
func FindLibTorsocks(filePath string) (string, error) {
	//Search in user config path
	f := filepath.Join(filePath, libTorsocks)
	if config.FileExists(f) {
		return f, nil
	}

	//Search in local directories
	for _, ld := range allLibDirs() {
		f = filepath.Join(ld, libTorsocks)
		if config.FileExists(f) {
			return f, nil
		}
	}

	//Search in bundle path
	pathCWD, err := os.Getwd()
	if err == nil {
		c := filepath.Join(pathCWD, "tor/")
		f = filepath.Join(c, libTorsocks)
		if config.FileExists(f) {
			return f, nil
		}
	}

	return "", errors.New("libtorsocks not found")
}
