package tor

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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
	IsPortAvailable(port int) bool
	GetRandomPort() int
}

type filepathFacade interface {
	Glob(string) ([]string, error)
}

type execFacade interface {
	LookPath(string) (string, error)
	ExecWithModify(bin string, args []string, cm ModifyCommand) ([]byte, error)
	StartCommand(*exec.Cmd) error
	WaitCommand(*exec.Cmd) error
}

type filesystemFacade interface {
	FileExists(string) bool
	IsADirectory(string) bool
	TempDir(where, suffix string) (string, error)
	EnsureDir(string, os.FileMode)
	WriteFile(string, []byte, os.FileMode) error
}

type torgoFacade interface {
	NewController(string) (torgoController, error)
}

type httpFacade interface {
	CheckConnectionOverTor(host string, port int) bool
	HTTPRequest(host string, port int, url string) (string, error)
}

var osf osFacade
var filepathf filepathFacade
var execf execFacade
var filesystemf filesystemFacade
var torgof torgoFacade
var httpf httpFacade

func setDefaultFacades() {
	osf = &realOsImplementation{}
	filepathf = &realFilepathImplementation{}
	execf = &realExecImplementation{}
	filesystemf = &realFilesystemImplementation{}
	torgof = &realTorgoImplementation{}
	httpf = &realHTTPImplementation{}
}

func init() {
	setDefaultFacades()
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

func (*realOsImplementation) IsPortAvailable(port int) bool {
	return config.IsPortAvailable(port)
}

func (*realOsImplementation) GetRandomPort() int {
	return config.GetRandomPort()
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
	// This executes the tor command, which is under control of the code
	/* #nosec G204 */
	cmd := exec.Command(bin, args...)

	if cm != nil {
		cm(cmd)
	}

	return cmd.Output()
}

func (*realExecImplementation) StartCommand(cmd *exec.Cmd) error {
	return cmd.Start()
}

func (*realExecImplementation) WaitCommand(cmd *exec.Cmd) error {
	return cmd.Wait()
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

func (*realFilesystemImplementation) TempDir(where, suffix string) (string, error) {
	return ioutil.TempDir(where, suffix)
}

func (*realFilesystemImplementation) EnsureDir(name string, mode os.FileMode) {
	config.EnsureDir(name, mode)
}

func (*realFilesystemImplementation) WriteFile(name string, content []byte, mode os.FileMode) error {
	return ioutil.WriteFile(name, content, mode)
}

type realTorgoImplementation struct{}

func (*realTorgoImplementation) NewController(a string) (torgoController, error) {
	return torgo.NewController(a)
}

type realHTTPImplementation struct{}

func (*realHTTPImplementation) CheckConnectionOverTor(host string, port int) bool {
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

func (*realHTTPImplementation) HTTPRequest(host string, port int, u string) (string, error) {
	proxyURL, err := url.Parse("socks5://" + net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return "", err
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return "", err
	}

	t := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: t}

	resp, err := client.Get(u)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid request")
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
