package tor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"autonomia.digital/tonio/app/config"
)

const defaultInstanceConfigDir = "tor"
const defaultInstanceConfigName = "torrc"
const defaultInstanceConfigData = "data"
const defaultInstanceSocksPort = 9950
const defaultInstanceControlPort = 9951
const defaultInstanceControlHost = "127.0.0.1"

// Instance if a representation of our Tor Control Port instance
type Instance struct {
	started       bool
	configFile    string
	socksPort     int
	controlHost   string
	controlPort   int
	dataDirectory string
	runningTor    *runningTor
	ioLock        sync.Mutex
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
func NewInstance() *Instance {
	i := createOurInstance()

	i.createConfigFile()

	return i
}

// Start our Tor Control Port
func (i *Instance) Start() error {
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

// Destroy close our instance running
func (i *Instance) Destroy() {
	d := filepath.Dir(i.configFile)
	if _, err := os.Stat(d); !os.IsNotExist(err) {
		_ = os.RemoveAll(d)
	}

	if i.runningTor != nil {
		i.runningTor.closeTorService()
	}
}

// GetSocksPort returns the custom Torsocks Port
func (i *Instance) GetSocksPort() int {
	return i.socksPort
}

// GetHost returns the custom Tor Control Port instance host
func (i *Instance) GetHost() string {
	return i.controlHost
}

// GetControlPort returns the custom Tor Control Port instance port
func (i *Instance) GetControlPort() int {
	return i.controlPort
}

func getTempDir() string {
	dir := filepath.Join(config.Dir(), fmt.Sprintf(".%s-%d", defaultInstanceConfigDir, os.Getpid()))
	return dir
}

func createOurInstance() *Instance {
	d := getTempDir()
	routePort, controlPort := findAvailableTorPorts()

	i := &Instance{
		started:       false,
		configFile:    filepath.Join(d, defaultInstanceConfigName),
		socksPort:     routePort,
		controlHost:   defaultInstanceControlHost,
		controlPort:   controlPort,
		dataDirectory: filepath.Join(d, defaultInstanceConfigData),
	}

	return i
}

func findAvailableTorPorts() (int, int) {
	controlPort := defaultInstanceControlPort
	if !config.IsPortAvailable(controlPort) {
		controlPort = config.GetRandomPort()
	}

	routePort := defaultInstanceSocksPort
	if !config.IsPortAvailable(routePort) {
		routePort = config.GetRandomPort()
	}

	return controlPort, routePort
}

func (i *Instance) createConfigFile() {
	i.ioLock.Lock()
	defer i.ioLock.Unlock()

	config.EnsureDir(filepath.Dir(i.configFile), 0700)
	config.EnsureDir(i.dataDirectory, 0700)

	go func() {
		err := i.writeToFile()
		if err != nil {
			log.Println(err)
		}
	}()
}

func (i *Instance) getConfigFileContents() []byte {
	content := []string{
		fmt.Sprintf("SocksPort %d", i.socksPort),
		fmt.Sprintf("ControlPort %d", i.controlPort),
		fmt.Sprintf("DataDirectory %s", i.dataDirectory),
		fmt.Sprintf("CookieAuthentication %d", 1),
	}
	return []byte(strings.Join(content, "\n"))
}

func (i *Instance) writeToFile() error {
	i.ioLock.Lock()
	defer i.ioLock.Unlock()
	return config.SafeWrite(i.configFile, i.getConfigFileContents(), 0600)
}

func (r *runningTor) closeTorService() {
	r.cancelFunc()
}

func (r *runningTor) waitForFinish() {
	e := r.cmd.Wait()
	r.finished = true
	r.finishedWithError = e
	r.finishChannel <- true
}
