package tor

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"autonomia.digital/tonio/app/config"
)

const torConfigName = "torrc"
const torConfigData = "data"
const defaultSocksPort = 9950
const defaultControlPort = 9951
const defaultControlHost = "127.0.0.1"

// Instance contains functions to work with Tor instance
type Instance interface {
	Start() error
	Destroy()
	GetController() Control
	GetHost() string
	GetControlPort() int
	GetRoutePort() int
	Exec(command string, args []string) (*RunningCommand, error)
}

type instance struct {
	started       bool
	configFile    string
	socksPort     int
	controlHost   string
	controlPort   int
	dataDirectory string
	controller    Control
	isLocal       bool
	runningTor    *runningTor
}

type runningTor struct {
	cmd               *exec.Cmd
	ctx               context.Context
	cancelFunc        context.CancelFunc
	finished          bool
	finishedWithError error
	finishChannel     chan bool
}

// GetSystem returns the Instance for working with Tor
func GetSystem() (Instance, error) {
	ensureTonioDataDir()

	conn := NewDefaultChecker()
	total, partial := conn.Check()

	if total != nil {
		return nil, errors.New("error: Tor is not available or supported in your system")
	}

	if partial != nil {
		return getOurInstance()
	}

	// TODO: We should check the local instance again?
	i := createSystemInstance()

	return i, nil
}

func getOurInstance() (Instance, error) {
	i, _ := NewInstance()

	err := i.Start()
	if err != nil {
		return nil, errors.New("error: we can't start our instance")
	}

	checker := NewChecker(false, i.GetHost(), i.GetRoutePort(), i.GetControlPort())

	timeout := time.Now().Add(10 * time.Second)
	for {
		total, partial := checker.Check()
		if total != nil {
			return nil, errors.New("error: we can't check our instance")
		}

		if time.Now().After(timeout) {
			return nil, errors.New("error: we can't start our instance")
		}

		if partial == nil {
			return i, nil
		}
	}
}

// NewInstance initialized our Tor Control Port instance
func NewInstance() (Instance, error) {
	i := createOurInstance()

	err := i.createConfigFile()

	return i, err
}

// Start our Tor Control Port
func (i *instance) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "tor", "-f", i.configFile)
	if err := cmd.Start(); err != nil {
		cancelFunc()
		return err
	}

	state := &runningTor{
		cmd:               cmd,
		ctx:               ctx,
		cancelFunc:        cancelFunc,
		finished:          false,
		finishedWithError: nil,
		finishChannel:     make(chan bool, 100),
	}

	i.started = true
	i.runningTor = state

	go state.waitForFinish()

	return nil
}

// GetController returns a controller for the instance `i`
func (i *instance) GetController() Control {
	if i.controller == nil {
		i.controller = CreateController(i.controlHost, i.controlPort, "")
	}
	return i.controller
}

// GetHost returns the instance host name
func (i *instance) GetHost() string {
	return i.controlHost
}

// GetControlPort returns the instance control port
func (i *instance) GetControlPort() int {
	return i.controlPort
}

// GetRoutePort returns the instance socket port
func (i *instance) GetRoutePort() int {
	return i.socksPort
}

// Destroy close our instance running
func (i *instance) Destroy() {
	_ = os.RemoveAll(filepath.Dir(i.configFile))

	if i.controller != nil {
		i.controller.DeleteOnionServices()
		i.controller = nil
	}

	if i.runningTor != nil {
		i.runningTor.closeTorService()
		i.runningTor = nil
	}
}

// RunningCommand is a representation of a torify command
type RunningCommand struct {
	Ctx        context.Context
	Cmd        *exec.Cmd
	CancelFunc context.CancelFunc
}

func (i *instance) Exec(command string, args []string) (*RunningCommand, error) {
	if i.isLocal {
		return i.torify(command, args)
	}
	return i.torsocks(command, args)
}

func (i *instance) torify(command string, args []string) (*RunningCommand, error) {
	arguments := append([]string{command}, args...)
	return i.exec("torify", arguments)
}

func (i *instance) torsocks(command string, args []string) (*RunningCommand, error) {
	arguments := append([]string{command}, args...)
	arguments = append(arguments, []string{
		"--address", i.controlHost,
		"--port", strconv.Itoa(i.socksPort),
	}...)

	return i.exec("torsocks", arguments)
}

func (i *instance) exec(command string, args []string) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command, args...)

	if err := cmd.Start(); err != nil {
		cancelFunc()
		return nil, err
	}

	rc := &RunningCommand{
		Ctx:        ctx,
		Cmd:        cmd,
		CancelFunc: cancelFunc,
	}

	return rc, nil
}

var tonioDataDir = filepath.Join(config.XdgDataHome(), "tonio")

func ensureTonioDataDir() {
	_ = os.MkdirAll(tonioDataDir, 0700)
}

func createOurInstance() *instance {
	d, _ := ioutil.TempDir(tonioDataDir, "tor")
	controlPort, routePort := findAvailableTorPorts()

	i := &instance{
		started:       false,
		configFile:    filepath.Join(d, torConfigName),
		controlHost:   defaultControlHost,
		controlPort:   controlPort,
		socksPort:     routePort,
		dataDirectory: filepath.Join(d, torConfigData),
		isLocal:       false,
		controller:    nil,
	}

	return i
}

func createSystemInstance() *instance {
	d, _ := ioutil.TempDir(tonioDataDir, "tor")

	i := &instance{
		started:       false,
		configFile:    filepath.Join(d, torConfigName),
		socksPort:     DefaultRoutePort,
		controlHost:   DefaultHost,
		controlPort:   DefaultControlPort,
		dataDirectory: filepath.Join(d, torConfigData),
		isLocal:       true,
		controller:    nil,
	}

	return i
}

func findAvailablePort(initial int) int {
	port := initial
	for !config.IsPortAvailable(port) {
		port = config.GetRandomPort()
	}
	return port
}

func findAvailableTorPorts() (controlPort, routePort int) {
	controlPort = findAvailablePort(defaultControlPort)
	routePort = findAvailablePort(defaultSocksPort)
	return
}

func (i *instance) createConfigFile() error {
	config.EnsureDir(i.dataDirectory, 0700)
	fmt.Printf("Saving the config file to: %s\n", i.configFile)
	return i.writeToFile()
}

func (i *instance) getConfigFileContents() []byte {
	noticeLog := filepath.Join(filepath.Dir(i.configFile), "notice.log")
	logFile := filepath.Join(filepath.Dir(i.configFile), "debug.log")

	content := fmt.Sprintf(
		`## Configuration file for a typical Tor user

## Tell Tor to open a SOCKS proxy on port %d
SOCKSPort %d

## Entry policies to allow/deny SOCKS requests based on IP address.
SOCKSPolicy reject *

## The port on which Tor will listen for local connections from Tor
## controller applications, as documented in control-spec.txt.
ControlPort %d

## The directory for keeping all the keys/etc.
DataDirectory %s

# Allow connections on the control port when the connecting process
# knows the contents of a file named "control_auth_cookie", which Tor
# will create in its data directory.
CookieAuthentication %d

## Send all messages of level 'notice' or higher to %s
Log notice file %s

## Send every possible message to %s
Log debug file %s`,
		i.socksPort,
		i.socksPort,
		i.controlPort,
		i.dataDirectory,
		0,
		noticeLog,
		noticeLog,
		logFile,
		logFile)
	return []byte(content)
}

func (i *instance) writeToFile() error {
	return ioutil.WriteFile(i.configFile, i.getConfigFileContents(), 0600)
}

func (r *runningTor) closeTorService() {
	r.cancelFunc()
}

func (r *runningTor) waitForFinish() {
	e := r.cmd.Wait()
	r.finished = true
	r.finishedWithError = e
	// TODO: Maybe here, we should check if the failure was because
	// of taken ports, regenerate the ports and try again?
	r.finishChannel <- true
}
