package tor

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"autonomia.digital/tonio/app/config"
)

var localPathTor = fmt.Sprintf("%s/%s", config.TorDir(), "tor")
var torBinaryPaths = []string{"/usr/bin/tor", "/usr/local/bin/tor", localPathTor}

// Binary contains functions to work with binary
type Binary interface {
	Check() error
	GetPaths() []string
	GetPathBinTor() string
	executeCmd(args []string) ([]byte, error)
}

type binary struct {
	paths      []string
	pathBinTor string
}

var torBinary *binary

// GetTorBinary returns the binary for working with it
func GetTorBinary(p []string) Binary {
	if torBinary == nil {
		initializeBinary(p)
	}
	return torBinary
}

// Check returned an error when no one binary were found
func (b *binary) Check() error {
	err := b.checkInPaths()
	if err != nil {
		return errors.New("no Tor binary found")
	}

	return nil
}

func initializeBinary(p []string) {
	torBinary = &binary{}
	if len(p) == 0 {
		torBinary.paths = torBinaryPaths
	} else {
		torBinary.paths = p
	}
}

func (b *binary) checkInPaths() error {
	for _, pathTor := range b.paths {
		b.pathBinTor = pathTor

		cmd := exec.Command(pathTor, "--version")
		err := cmd.Run()
		if err == nil {
			if torBinary.checkTorVersionCompatibility() {
				return nil
			}
		}
	}

	b.pathBinTor = ""
	return errors.New("no Tor binary installed")
}

// GetPathBinTor return the path to Tor binary
func (b *binary) GetPathBinTor() string {
	return b.pathBinTor
}

func (b *binary) GetPaths() []string {
	return b.paths
}

func (b *binary) executeCmd(args []string) ([]byte, error) {
	p := torBinary.pathBinTor
	cmd := exec.Command(p, args...)
	output, err := cmd.Output()
	if output == nil || err != nil {
		return nil, errors.New("invalid command")
	}
	return output, nil
}

func (b *binary) checkTorVersionCompatibility() bool {
	output, err := torBinary.executeCmd([]string{"--version"})
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
