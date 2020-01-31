package tor

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	Exec(command string, args []string, env []string) (*RunningCommand, error)
	GetPathBinary() string
	SetPathBinary(path string)
	GetBundled() bool
	SetBundled(bool)
}

type instance struct {
	started       bool
	configFile    string
	socksPort     int
	controlHost   string
	controlPort   int
	dataDirectory string
	password      string
	useCookie     bool
	controller    Control
	isLocal       bool
	runningTor    *runningTor
	pathBinary    string
	pathTorsocks  string
	isBundled     bool
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
func GetSystem(conf *config.ApplicationConfig) (Instance, error) {
	i, err := getSystemInstance()
	if err == nil {
		return i, nil
	}

	ensureTonioDataDir()

	binaryPath, isBundled := Initialize(conf.GetPathTor())
	if len(binaryPath) == 0 {
		return nil, errors.New("error: there is no valid or available Tor binary")
	}

	log.Printf("Using Tor binary found in: %s", binaryPath)

	i, err = getOurInstance(binaryPath, conf.GetPathTorSocks(), isBundled)
	if err != nil {
		log.Fatal(err)
	}

	return i, nil
}

const torStartupTimeout = 2 * time.Minute

func getSystemInstance() (Instance, error) {
	checker := NewDefaultChecker()

	total, partial := checker.Check()

	if total != nil || partial != nil {
		return nil, errors.New("error: we can't use the system instance")
	}

	i := &instance{
		started:     true,
		controlHost: *config.TorHost,
		controlPort: *config.TorPort,
		socksPort:   *config.TorRoutePort,
		password:    *config.TorControlPassword,
		isLocal:     true,
		isBundled:   false,
		controller:  nil,
		pathBinary:  "tor",
	}

	return i, nil
}

func getOurInstance(binaryPath string, torsocksPath string, isBundled bool) (Instance, error) {
	i, _ := NewInstance(binaryPath, torsocksPath)
	i.SetBundled(isBundled)

	err := i.Start()
	if err != nil {
		return nil, errors.New("error: we can't start our instance")
	}

	h := i.GetHost()
	s := i.GetRoutePort()
	c := i.GetControlPort()

	checker := NewChecker(true, h, s, c, "")

	timeout := time.Now().Add(torStartupTimeout)
	for {
		time.Sleep(3 * time.Second)
		total, partial := checker.Check()
		if total != nil {
			return nil, errors.New("error: we can't check our instance")
		}

		if time.Now().After(timeout) {
			return nil, errors.New("error: we can't start our instance because timeout")
		}

		if partial == nil {
			return i, nil
		}
	}
}

// NewInstance initialized our Tor Control Port instance
func NewInstance(pathBinary string, torsocksPath string) (Instance, error) {
	i := createOurInstance(pathBinary, torsocksPath)

	err := i.createConfigFile()

	return i, err
}

// Start our Tor Control Port
func (i *instance) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, i.pathBinary, "-f", i.configFile)

	if i.isBundled {
		log.Println("Tor isBundled..")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=.")
	}

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
		i.controller = CreateController(i.controlHost, i.controlPort)

		if len(i.password) != 0 {
			i.controller.SetPassword(i.password)
		}

		if i.useCookie {
			i.controller.UseCookieAuth()
		}
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

// GetPathBinary return the value of the property pathBinary
func (i *instance) GetPathBinary() string {
	return i.pathBinary
}

// SetPathBinary set the value for the property pathBinary
func (i *instance) SetPathBinary(path string) {
	i.pathBinary = path
}

func (i *instance) GetBundled() bool {
	return i.isBundled
}

func (i *instance) SetBundled(bundled bool) {
	i.isBundled = bundled
}

// RunningCommand is a representation of a torify command
type RunningCommand struct {
	Ctx        context.Context
	Cmd        *exec.Cmd
	CancelFunc context.CancelFunc
}

func (i *instance) Exec(command string, args []string, env []string) (*RunningCommand, error) {
	return i.runOurTorsocks(command, args, env)
}

func (i *instance) runOurTorsocks(command string, args []string, env []string) (*RunningCommand, error) {
	return i.exec(command, args, i.isBundled, env)
}

func (i *instance) exec(command string, args []string, libtorsocks bool, env []string) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command, args...)

	if libtorsocks {
		pathTorsocks, err := FindLibTorsocks(i.pathTorsocks)
		if err != nil {
			cancelFunc()
			return nil, errors.New("error: libtorsocks.so was not found")
		}

		pwd := [32]byte{}
		_ = config.RandomString(pwd[:])

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("LD_PRELOAD=%s", pathTorsocks))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TORSOCKS_PASSWORD=%s", string(pwd[:])))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TORSOCKS_TOR_ADDRESS=%s", i.controlHost))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TORSOCKS_TOR_PORT=%d", i.socksPort))
	}

	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}

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

func createOurInstance(pathBinary string, torsocksPath string) *instance {
	d, _ := ioutil.TempDir(tonioDataDir, "tor")
	controlPort, routePort := findAvailableTorPorts()

	i := &instance{
		started:       false,
		configFile:    filepath.Join(d, torConfigName),
		controlHost:   defaultControlHost,
		controlPort:   controlPort,
		socksPort:     routePort,
		dataDirectory: filepath.Join(d, torConfigData),
		password:      "", // our instance don't use authentication with password
		useCookie:     true,
		isLocal:       false,
		controller:    nil,
		pathBinary:    pathBinary,
		pathTorsocks:  torsocksPath,
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
	log.Printf("Saving the config file to: %s\n", i.configFile)
	return i.writeToFile()
}

func (i *instance) getConfigFileContents() []byte {
	noticeLog := filepath.Join(filepath.Dir(i.configFile), "notice.log")
	logFile := filepath.Join(filepath.Dir(i.configFile), "debug.log")

	cookieFile := 1
	if !i.useCookie {
		cookieFile = 0
	}

	content := fmt.Sprintf(
		`## Configuration file for a typical Tor user

## Tell Tor to open a SOCKS proxy on port %d
SOCKSPort %d

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
		cookieFile,
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
