package tor

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	"github.com/digitalautonomy/wahay/config"
)

const (
	torConfigName      = "torrc"
	torConfigData      = "data"
	defaultSocksPort   = 9050
	defaultControlPort = 9051
	defaultControlHost = "127.0.0.1"
)

// Instance contains functions to work with Tor instance
type Instance interface {
	Start() error
	Destroy()
	GetController() Control
	Exec(command string, args []string, pre ModifyCommand) (*RunningCommand, error)
	HTTPrequest(url string) (string, error)
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
	isLocal       bool
	pathTorsocks  string
	controller    Control
	runningTor    *runningTor
	binary        *binary
}

// TODO[OB] - This design is _very_ specific to the needs of the certificate getter
// It would be better if the instance could return a dialer - then anyone/anything
// could easily use it to do Tor network traffic, and this HTTP specific stuff
// could be done in the certificate package

func (i *instance) HTTPrequest(u string) (string, error) {
	proxyURL, err := url.Parse("socks5://" + net.JoinHostPort(i.controlHost, strconv.Itoa(i.socksPort)))
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

type runningTor struct {
	cmd               *exec.Cmd
	ctx               context.Context
	cancelFunc        context.CancelFunc
	finished          bool
	finishedWithError error
	finishChannel     chan bool
}

var currentInstance Instance

func setSingleInstance(i Instance) {
	currentInstance = i
}

// TODO[OB] - Lots of getters here...

func getSingleInstance() (Instance, error) {
	if currentInstance == nil {
		return nil, errors.New("no instance initialized")
	}
	return currentInstance, nil
}

// CurrentInstance returns the current Tor instance
func CurrentInstance() Instance {
	return currentInstance
}

// GetController returns the Tor controller for the current instance
func GetController() (Control, error) {
	i := CurrentInstance()

	if i == nil {
		return nil, errors.New("tor hasn't been initialized")
	}

	return i.GetController(), nil
}

// TODO[OB] - Why is this function exposed? It isn't used anywhere outside

// DeleteOnionService deletes a specific ONION service for the
// current Tor instance controller
func DeleteOnionService(serviceID string) error {
	controller, err := GetController()
	if err != nil {
		return err
	}

	return controller.DeleteOnionService(serviceID)
}

// Onion is a representation of a Tor Onion Service
type Onion interface {
	ID() string
	Delete() error
}

type onion struct {
	id    string
	ports []OnionPort
}

func (s *onion) ID() string {
	return s.id
}

func (s *onion) Delete() error {
	return DeleteOnionService(s.id)
}

// NewOnionServiceWithMultiplePorts creates a new Onion service for the current Tor controller
func NewOnionServiceWithMultiplePorts(ports []OnionPort) (Onion, error) {
	controller, err := GetController()
	if err != nil {
		return nil, err
	}

	serviceID, err := controller.CreateNewOnionServiceWithMultiplePorts(ports)
	if err != nil {
		return nil, err
	}

	s := &onion{
		id:    serviceID,
		ports: ports,
	}

	return s, nil
}

var (
	// ErrTorBinaryNotFound is an error to be trown when wasn't
	// possible to find any available or valid Tor binary
	ErrTorBinaryNotFound = errors.New("no Tor binary found")

	// ErrTorInstanceCantStart is an error to be trown when the
	// Tor instance cannot be started
	ErrTorInstanceCantStart = errors.New("the Tor instance cannot start")

	// ErrTorConnectionTimeout is an error to be trown when the
	// connection to the Tor network using our instance wasn't possible
	ErrTorConnectionTimeout = errors.New("connection over Tor timeout")
)

// GetInstance returns the Instance for working with Tor
// This function should be called only once during the system initialization
func GetInstance(conf *config.ApplicationConfig) (Instance, error) {
	i, err := getSingleInstance()
	if err == nil {
		return i, nil
	}

	// Checking if the system Tor can be used.
	// This should work for system like Tails, where Tor is
	// already available in the system.
	i, err = systemInstance()
	if err == nil {
		setSingleInstance(i)
		return i, nil
	}

	// Proceed to initialize our Tor instance
	ensureWahayDataDir()

	b, err := findTorBinary(conf)
	if b == nil || err != nil {
		return nil, ErrTorBinaryNotFound
	}

	log.Printf("Using Tor binary found in: %s", b.path)

	i, err = getOurInstance(b, conf)
	if err != nil {
		log.Debugf("tor.GetInstance() error: %s", err)
		return nil, err
	}

	setSingleInstance(i)

	return i, nil
}

const torStartupTimeout = 2 * time.Minute

func systemInstance() (Instance, error) {
	checker := newDefaultChecker()

	total, partial := checker.Check()

	if total != nil || partial != nil {
		return nil, errors.New("error: we can't use system Tor instance")
	}

	i := &instance{
		started:     true,
		controlHost: defaultControlHost,
		controlPort: defaultControlPort,
		socksPort:   defaultSocksPort,
		useCookie:   false,
		isLocal:     true,
	}

	return i, nil
}

func getOurInstance(b *binary, conf *config.ApplicationConfig) (*instance, error) {
	i, _ := newInstance(b, conf.GetPathTorSocks())

	err := i.Start()
	if err != nil {
		return nil, err
	}

	checker := newCustomChecker(i.controlHost, i.socksPort, i.controlPort)

	timeout := time.Now().Add(torStartupTimeout)
	for {
		time.Sleep(3 * time.Second)

		errTotal, errPartial := checker.Check()
		if errTotal != nil {
			return nil, errTotal
		}

		if time.Now().After(timeout) {
			return nil, ErrTorConnectionTimeout
		}

		if errPartial == nil {
			return i, nil
		}

		log.WithFields(log.Fields{
			"time": time.Now(),
		}).Error(fmt.Sprintf("The following error occurred while checking Tor connectivity: %s", errPartial.Error()))
	}
}

func newInstance(b *binary, torsocksPath string) (*instance, error) {
	i := createOurInstance(b, torsocksPath)

	err := i.createConfigFile()

	return i, err
}

// Start our Tor Control Port
func (i *instance) Start() error {
	if i.binary == nil || !i.binary.isValid {
		return ErrTorInstanceCantStart
	}

	state, err := i.binary.start(i.configFile)
	if err != nil {
		return err
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

// Destroy close our instance running
func (i *instance) Destroy() {
	log.Debugf("Removing custom Tor temp dir: %s", filepath.Dir(i.configFile))
	err := os.RemoveAll(filepath.Dir(i.configFile))
	if err != nil {
		log.Debug(err)
	}

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

func (i *instance) Exec(command string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	return i.runOurTorsocks(command, args, pre)
}

func (i *instance) runOurTorsocks(command string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	return i.exec(command, args, pre)
}

// ModifyCommand is a function that will potentially modify a command
type ModifyCommand func(*exec.Cmd)

func (i *instance) exec(command string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command, args...)

	pathTorsocks, err := findLibTorsocks(i.pathTorsocks)
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

	if pre != nil {
		pre(cmd)
	}

	if *config.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
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

var wahayDataDir = filepath.Join(config.XdgDataHome(), "wahay")

func ensureWahayDataDir() {
	_ = os.MkdirAll(wahayDataDir, 0700)
}

func createOurInstance(b *binary, torsocksPath string) *instance {
	d, _ := ioutil.TempDir(wahayDataDir, "tor")

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
		pathTorsocks:  torsocksPath,
		binary:        b,
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

	// TODO[OB] - This should probably be exported into
	// its own file, and then use esc to include it

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
