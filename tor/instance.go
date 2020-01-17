package tor

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

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
}

type instance struct {
	started       bool
	configFile    string
	socksPort     int
	controlHost   string
	controlPort   int
	dataDirectory string
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

// Destroy close our instance running
func (i *instance) Destroy() {
	_ = os.RemoveAll(filepath.Dir(i.configFile))

	if i.runningTor != nil {
		i.runningTor.closeTorService()
		i.runningTor = nil
	}
}

// GetSocksPort returns the custom Torsocks Port
func (i *instance) GetSocksPort() int {
	return i.socksPort
}

// GetHost returns the custom Tor Control Port instance host
func (i *instance) GetHost() string {
	return i.controlHost
}

// GetControlPort returns the custom Tor Control Port instance port
func (i *instance) GetControlPort() int {
	return i.controlPort
}

var tonioDataDir = filepath.Join(config.XdgDataHome(), "tonio")

func ensureTonioDataDir() {
	os.MkdirAll(tonioDataDir, 0700)
}

func createOurInstance() *instance {
	ensureTonioDataDir()
	d, _ := ioutil.TempDir(tonioDataDir, "tor")
	controlPort, routePort := findAvailableTorPorts()

	i := &instance{
		started:       false,
		configFile:    filepath.Join(d, torConfigName),
		socksPort:     routePort,
		controlHost:   defaultControlHost,
		controlPort:   controlPort,
		dataDirectory: filepath.Join(d, torConfigData),
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
	return i.writeToFile()
}

func (i *instance) getConfigFileContents() []byte {
	content := fmt.Sprintf(`
SocksPort %d
ControlPort %d
DataDirectory %s
CookieAuthentication %d`,
		i.socksPort, i.controlPort, i.dataDirectory, 1)
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
