package tor

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"autonomia.digital/tonio/app/config"
)

// Instance if a representation of our Tor Control Port instance
type Instance struct {
	started        bool
	configFile     string
	configFileName string
	socksPort      int
	controlHost    string
	controlPort    int
	dataDirectory  string
	authType       string
	runningTor     *runningTor
	ioLock         sync.Mutex
}

type runningTor struct {
	cmd               *exec.Cmd
	ctx               context.Context
	cancelFunc        context.CancelFunc
	finished          bool
	finishedWithError error
	finishChannel     chan bool
}

func (i *Instance) close() {
	if i.runningTor != nil {
		i.runningTor.close()
	}
}

func (r *runningTor) close() {
	r.cancelFunc()
}

func (r *runningTor) waitForFinish() {
	e := r.cmd.Wait()
	r.finished = true
	r.finishedWithError = e
	r.finishChannel <- true
}

// NewInstance initialized our Tor Control Port instance
func NewInstance() (*Instance, error) {
	log.Println("Creating new Tor Control Port instance")

	i := &Instance{
		started:        false,
		configFile:     "",
		configFileName: "tor/torrc",
		socksPort:      9950,
		controlHost:    "127.0.0.1",
		controlPort:    9951,
		dataDirectory:  "",
		authType:       "cookie",
	}

	err := i.loadOrCreateConfigFile()

	return i, err
}

// GetHost returns the custom Tor Control Port instance host
func (i *Instance) GetHost() string {
	return i.controlHost
}

// GetControlPort returns the custom Tor Control Port instance port
func (i *Instance) GetControlPort() string {
	return strconv.Itoa(i.controlPort)
}

// GetPreferredAuthType returns the custom Tor Control Port authentication mode
func (i *Instance) GetPreferredAuthType() string {
	return i.authType
}

// Start our Tor Control Port
func (i *Instance) Start() error {
	log.Println("Starting our Tor Control Port instance")

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

func (i *Instance) loadOrCreateConfigFile() error {
	i.ioLock.Lock()
	defer i.ioLock.Unlock()

	i.configFile = config.FindFile(i.configFileName, "")
	i.dataDirectory = filepath.Join(filepath.Dir(i.configFile), "data")
	config.EnsureDir(i.dataDirectory, 0700)

	data, exists, err := i.loadConfigFile()
	if exists && err != nil {
		return err
	}

	if !exists || len(data) == 0 {
		go func() {
			err = i.save()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	return nil
}

func (i *Instance) save() error {
	i.ioLock.Lock()
	defer i.ioLock.Unlock()

	return config.SafeWrite(i.configFile, i.getConfigFileContents(), 0600)
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

func (i *Instance) loadConfigFile() ([]byte, bool, error) {
	if config.FileExists(i.configFile) {
		data, err := ioutil.ReadFile(i.configFile)
		return data, true, err
	}
	return nil, false, errors.New("tor control port file not found")
}
