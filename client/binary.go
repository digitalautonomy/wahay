package client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

var (
	errInvalidCommand             = errors.New("invalid command")
	errInvalidBinaryFile          = errors.New("the defined binary file don't exists")
	errBinaryAlreadyExists        = errors.New("the binary already exists in the destination directory")
	errDestinationIsNotADirectory = errors.New("the destination to copy the binary is not a directory")
	errNoClientInConfiguredPath   = errors.New("no client in the configured path")
)

const (
	mumbleBundleLibsDir   = "lib"
	mumbleBundlePath      = "mumble/mumble"
	wahayMumbleBundlePath = "wahay/mumble/mumble"
)

type binary struct {
	path           string
	env            []string
	isValid        bool
	isBundle       bool
	lastError      error
	isTemporary    bool
	shouldBeCopied bool
}

func (b *binary) getPath() string {
	return b.path
}

func (b *binary) getEnv() []string {
	if !b.isBundle {
		return nil
	}

	if len(b.env) == 0 {
		isBundle, env := checkLibsDependenciesInPath(b.path)
		if !isBundle || len(env) == 0 {
			b.isBundle = false
			return nil
		}
	}

	return b.env
}

func (b *binary) binaryExists() bool {
	return fileExists(b.path)
}

func (b *binary) copyTo(path string) error {
	if !b.isValid || !b.binaryExists() {
		return errInvalidBinaryFile
	}

	if len(path) == 0 || !isADirectory(path) {
		return errDestinationIsNotADirectory
	}

	mumbleCopyFile := filepath.Join(path, "mumble")

	if fileExists(mumbleCopyFile) {
		return errBinaryAlreadyExists
	}

	err := b.copyBinaryToDir(mumbleCopyFile)
	if err != nil {
		return errInvalidBinaryFile
	}

	b.path = filepath.Join(mumbleCopyFile)
	b.isTemporary = true

	return nil
}

func (b *binary) cleanup() {
	b.remove()
}

func (b *binary) shouldBeRemoved() bool {
	return b.isTemporary
}

func (b *binary) remove() {
	if b.shouldBeRemoved() {
		err := os.RemoveAll(filepath.Dir(b.path))
		if err != nil {
			log.Errorf("An error occurred while removing Mumble temp directory: %s", err.Error())
		}
	}
}

func (b *binary) copyBinaryToDir(destination string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(b.path); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(destination); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}

	if srcinfo, err = os.Stat(b.path); err != nil {
		return err
	}

	return os.Chmod(destination, srcinfo.Mode())
}

func newMumbleBinary(path string) *binary {
	b := &binary{
		isValid:        true,
		isBundle:       false,
		env:            []string{},
		lastError:      nil,
		shouldBeCopied: false,
		isTemporary:    false,
	}

	p, err := getRealMumbleBinaryPath(path)
	if len(p) == 0 || err != nil {
		b.isValid = false
		return b
	}

	b.path = p

	return b
}

func getRealMumbleBinaryPath(path string) (string, error) {
	if len(path) == 0 {
		return "", errors.New("invalid binary path")
	}

	if isADirectory(path) {
		// TODO: should we find all the Mumble binary possibilities inside the directory?
		// Examples:
		// 	 - mumble
		//   - mumble-0.1.0.4
		//   - mumble-beta
		//   - mumble-bin
		return filepath.Join(path, mumbleBundlePath), nil
	}

	return path, nil
}

func getMumbleBinary(conf *config.ApplicationConfig) *binary {
	callbacks := []func() (*binary, error){
		getMumbleBinaryInConf(conf),
		getMumbleBinaryInLocalDir,
		getMumbleBinaryInCurrentWorkingDir,
		getMumbleBinaryInDataDir,
		getMumbleBinaryInSystem,
	}

	for _, getBinary := range callbacks {
		b, err := getBinary()

		if err != nil {
			log.Debugf("Mumble binary error: %s", err)
			break
		}

		if b == nil {
			log.Debugf("Mumble binary error: Not found")
			continue
		}

		if b.lastError != nil {
			log.Debugf("Mumble binary error: %s", b.lastError)
			continue
		}

		if !b.isValid {
			continue
		}

		return b
	}

	return nil
}

func getMumbleBinaryInConf(conf *config.ApplicationConfig) func() (*binary, error) {
	return func() (*binary, error) {
		b := isAnAvailableMumbleBinary(conf.GetPathMumble())
		if b == nil || b.lastError != nil {
			return nil, errNoClientInConfiguredPath
		}

		return b, nil
	}
}

func getMumbleBinaryInLocalDir() (*binary, error) {
	localDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, nil
	}

	b := isAnAvailableMumbleBinary(filepath.Join(localDir, mumbleBundlePath))

	return b, nil
}

func getMumbleBinaryInCurrentWorkingDir() (*binary, error) {
	cwDir, err := os.Getwd()
	if err != nil {
		return nil, nil
	}

	b := isAnAvailableMumbleBinary(filepath.Join(cwDir, mumbleBundlePath))

	return b, nil
}

func getMumbleBinaryInDataDir() (*binary, error) {
	dataDir := config.XdgDataHome()
	dirs := []string{
		filepath.Join(dataDir, mumbleBundlePath),
		filepath.Join(dataDir, wahayMumbleBundlePath),
	}

	for _, d := range dirs {
		b := isAnAvailableMumbleBinary(d)
		if b != nil && b.isValid {
			return b, nil
		}
	}

	return nil, nil
}

func getMumbleBinaryInSystem() (*binary, error) {
	path, err := exec.LookPath("mumble")
	if err != nil {
		return nil, nil
	}

	b := isAnAvailableMumbleBinary(path)

	return b, nil
}

func isAnAvailableMumbleBinary(path string) *binary {
	log.Debugf("Checking Mumble binary in: <%s>", path)

	b := newMumbleBinary(path)
	if !b.isValid {
		return b
	}

	bin := b.path
	command := exec.Command(bin, "-h")

	isBundle, env := checkLibsDependenciesInPath(b.path)
	if isBundle && len(env) > 0 {
		command.Env = append(os.Environ(), env...)
		b.env = append(b.env, env...)
	}

	b.isBundle = isBundle
	b.shouldBeCopied = !isBundle

	output, err := command.Output()
	if len(output) == 0 && err != nil {
		b.isValid = false
		b.lastError = errInvalidCommand
		return b
	}

	return b
}

func checkLibsDependenciesInPath(path string) (isBundle bool, env []string) {
	libsDir := filepath.Join(filepath.Dir(path), mumbleBundleLibsDir)

	if directoryExists(libsDir) {
		env = append(env, fmt.Sprintf("LD_LIBRARY_PATH=%s", libsDir))
		isBundle = true
	}

	return
}
