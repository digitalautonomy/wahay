package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	errInvalidCommand             = errors.New("invalid command")
	errInvalidBinaryFile          = errors.New("the defined binary file don't exists")
	errBinaryAlreadyExists        = errors.New("the binary already exists in the destination directory")
	errDestinationIsNotADirectory = errors.New("the destination to copy the binary is not a directory")
)

const (
	mumbleBundleLibsDir = "lib"
	mumbleBundlePath    = "mumble/mumble"
)

// Binary is a representation of the Mumble binary
type Binary interface {
	GetPath() string
	GetEnv() []string
	ShouldBeCopied() bool
	CopyTo(path string) error
	BinaryExists() bool
	IsValid() bool
	Cleanup()
}

type binary struct {
	sync.Mutex
	path           string
	env            []string
	isValid        bool
	isBundle       bool
	lastError      error
	isTemporary    bool
	shouldBeCopied bool
}

func (b *binary) GetPath() string {
	return b.path
}

func (b *binary) GetEnv() []string {
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

func (b *binary) ShouldBeCopied() bool {
	return b.shouldBeCopied
}

func (b *binary) BinaryExists() bool {
	return fileExists(b.path)
}

func (b *binary) CopyTo(path string) error {
	if !b.isValid || !b.BinaryExists() {
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

func (b *binary) IsValid() bool {
	return b.isValid
}

func (b *binary) Cleanup() {
	b.remove()
}

func (b *binary) shouldBeRemoved() bool {
	return b.isTemporary
}

func (b *binary) remove() {
	if b.shouldBeRemoved() {
		err := os.RemoveAll(filepath.Dir(b.path))
		if err != nil {
			log.Printf("An error occurred while removing Mumble temp directory: %s", err.Error())
		}
	}
}

func (b *binary) copyBinaryToDir(destination string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	b.Lock()
	defer b.Unlock()

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
		path:           path,
		isValid:        false,
		isBundle:       false,
		env:            []string{},
		lastError:      nil,
		shouldBeCopied: false,
		isTemporary:    false,
	}

	return b
}

func getMumbleBinary(userConfiguredPath string) Binary {
	binaries := []func() *binary{
		getMumbleBinaryInConf(userConfiguredPath),
		getMumbleBinaryInLocalDir,
		getMumbleBinaryInDataDir,
		getMumbleBinaryInSystem,
	}

	for _, getBinary := range binaries {
		b := getBinary()

		if b == nil {
			log.Printf("Mumble binary error: Not found")
			continue
		}

		if b.lastError != nil {
			log.Printf("Mumble binary error: %s", b.lastError)
			continue
		}

		return b
	}

	return nil
}

func getMumbleBinaryInConf(path string) func() *binary {
	return func() *binary {
		return isAnAvailableMumbleBinary(path)
	}
}

func getMumbleBinaryInLocalDir() *binary {
	localDir, err := os.Getwd()
	if err != nil {
		return nil
	}

	return isAnAvailableMumbleBinary(filepath.Join(localDir, mumbleBundlePath))
}

func getMumbleBinaryInDataDir() *binary {
	return isAnAvailableMumbleBinary(filepath.Join(dataHomeDir(), mumbleBundlePath))
}

func getMumbleBinaryInSystem() *binary {
	path, err := exec.LookPath("mumble")
	if err != nil {
		return nil
	}

	return isAnAvailableMumbleBinary(path)
}

func isAnAvailableMumbleBinary(path string) *binary {
	log.Printf("Checking Mumble binary in: <%s>", path)

	b := newMumbleBinary(path)

	if len(path) == 0 || !fileExists(path) {
		b.lastError = fmt.Errorf("the Mumble binary path is invalid or do not exists")
		return b
	}

	command := exec.Command(path, "-h")

	isBundle, env := checkLibsDependenciesInPath(path)
	if isBundle && len(env) > 0 {
		command.Env = append(os.Environ(), env...)
		b.env = append(b.env, env...)
	}

	b.isBundle = isBundle
	b.shouldBeCopied = !isBundle

	output, err := command.Output()
	if len(output) == 0 && err != nil {
		b.lastError = errInvalidCommand
		return b
	}

	b.isValid = true

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
