package tor

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/digitalautonomy/wahay/config"
)

// This file contains internal interfaces used to abstract
// away all external usage, so that we can effectively manage
// this from tests. Most important here are the different
// execution functionalities. The things here
// should not be exposed to the world. They are internal
// implementation details.

type osFacade interface {
	Getwd() (string, error)
	Args() []string
	Environ() []string
	RemoveAll(string) error
	MkdirAll(string, os.FileMode) error
	Stdout() *os.File
	Stderr() *os.File
}

type filepathFacade interface {
	Glob(string) ([]string, error)
}

type execFacade interface {
	LookPath(string) (string, error)
	ExecWithModify(bin string, args []string, cm ModifyCommand) ([]byte, error)
	StartCommand(*exec.Cmd) error
}

type filesystemFacade interface {
	FileExists(string) bool
	IsADirectory(string) bool
}

var osf osFacade
var filepathf filepathFacade
var execf execFacade
var filesystemf filesystemFacade

func init() {
	osf = &realOsImplementation{}
	filepathf = &realFilepathImplementation{}
	execf = &realExecImplementation{}
	filesystemf = &realFilesystemImplementation{}
}

type realOsImplementation struct{}

func (*realOsImplementation) Getwd() (string, error) {
	return os.Getwd()
}

func (*realOsImplementation) Args() []string {
	return os.Args
}

func (*realOsImplementation) Environ() []string {
	return os.Environ()
}

func (*realOsImplementation) RemoveAll(dir string) error {
	return os.RemoveAll(dir)
}

func (*realOsImplementation) MkdirAll(dir string, mode os.FileMode) error {
	return os.MkdirAll(dir, mode)
}

func (*realOsImplementation) Stdout() *os.File {
	return os.Stdout
}

func (*realOsImplementation) Stderr() *os.File {
	return os.Stderr
}

type realFilepathImplementation struct{}

func (*realFilepathImplementation) Glob(p string) ([]string, error) {
	return filepath.Glob(p)
}

type realExecImplementation struct{}

func (*realExecImplementation) LookPath(s string) (string, error) {
	return exec.LookPath(s)
}

func (*realExecImplementation) ExecWithModify(bin string, args []string, cm ModifyCommand) ([]byte, error) {
	cmd := exec.Command(bin, args...)

	if cm != nil {
		cm(cmd)
	}

	return cmd.Output()
}

func (*realExecImplementation) StartCommand(cmd *exec.Cmd) error {
	return cmd.Start()
}

type realFilesystemImplementation struct{}

func (*realFilesystemImplementation) FileExists(path string) bool {
	return config.FileExists(path)
}

func (*realFilesystemImplementation) IsADirectory(path string) bool {
	dir, err := os.Stat(path)
	if err != nil {
		return false
	}

	return dir.IsDir()
}
