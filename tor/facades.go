package tor

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/digitalautonomy/wahay/config"
	"github.com/wybiral/torgo"
	"golang.org/x/net/proxy"
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

type torgoFacade interface {
	NewController(string) (torgoController, error)
}

type httpFacade interface {
	CheckConnectionOverTor(host string, port int) bool
}

var osf osFacade
var filepathf filepathFacade
var execf execFacade
var filesystemf filesystemFacade
var torgof torgoFacade
var httpf httpFacade

func init() {
	osf = &realOsImplementation{}
	filepathf = &realFilepathImplementation{}
	execf = &realExecImplementation{}
	filesystemf = &realFilesystemImplementation{}
	torgof = &realTorgoImplementation{}
	httpf = &realHttpImplementation{}
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

type realTorgoImplementation struct{}

func (*realTorgoImplementation) NewController(a string) (torgoController, error) {
	return torgo.NewController(a)
}

type realHttpImplementation struct{}

func (*realHttpImplementation) CheckConnectionOverTor(host string, port int) bool {
	proxyURL, err := url.Parse("socks5://" + net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return false
	}

	t := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: t}

	resp, err := client.Get("https://check.torproject.org/api/ip")
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	var v checkTorResult
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return false
	}

	return v.IsTor
}
