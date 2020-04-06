package tor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/digitalautonomy/wahay/config"
	"github.com/wybiral/torgo"
	. "gopkg.in/check.v1"

	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
)

type TorAcceptanceSuite struct{}

var _ = Suite(&TorAcceptanceSuite{})

// These tests will try to document the expected behavior
// with regard to what Tor instance will be used or not used
// depending on what the system figures out about what
// is going on

// Specifically, a method should be called ONCE
// at system start to initialize the Tor subsystem. This method
// will first try to detect circumstances of the system it's
// running on, and then based on that create an instance that
// will be used for all subsequent Tor connections.

// The rules look more or less like this:
//   If Tor is available and running already:
//     - Test authentication method NONE
//     - Test authentication method COOKIE
//     - Test authentication method PASSWORD
//   If any of the authentication methods succeed:
//     - Check that the version of Tor is acceptable
//     - Check that Tor is actually connected to the
//       internet and can be used to do connections
//
//   If the System Tor is is acceptable, use it, with the
//   detected authentication method. In this case, do
//   not try to create a configuration file or data dir
//   for Tor. Also, do not try to stop it at the end of
//   Wahay.
//
//   If the System Tor is not possible to use, check whether
//   the Tor binary is available and acceptable (version is correct)
//
//   If the System Tor is not acceptable or not available, try to
//   find another Tor executable that can be used.
//
//   If no acceptable Tor executable is found, we have to give up
//
//   If an acceptable Tor executable is found, create a new data dir
//   and configuration file. Start Tor with this.
//   Run checks to make sure it's acceptable. If not, we have to give up.
//
//   At the end of Wahay, when running a custom Tor instance, stop the Tor
//   instance. Also clean up and destroy the created data directory

const systemTorBinary = "/usr/sbin/tor"

func (s *TorAcceptanceSuite) Test_thatSystemTorWillBeUsed_whenSystemTorIsOKWithNoAuthenticationAndProperVersion(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	tc := &mockTorgoController{}
	tc.authNoneReturn = nil
	tc.authPassReturn = errors.New("couldn't auth")
	tc.authCookieReturn = errors.New("couldn't auth")
	tc.getVersionReturn1 = "4.0.1"
	tc.getVersionReturn2 = nil

	mocktorgof.newControllerReturn1 = tc

	mockhttpf.checkConnectionReturn = true

	ix, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, IsNil)

	c.Assert(mocktorgof.newControllerArg, Equals, "127.0.0.1:9051")

	c.Assert(tc.authNoneCalled, Equals, 2)
	c.Assert(tc.authPassCalled, Equals, 0)
	c.Assert(tc.authCookieCalled, Equals, 0)

	c.Assert(tc.getVersionCalled, Equals, 1)

	c.Assert(mockhttpf.checkConnectionArg1, Equals, "127.0.0.1")
	c.Assert(mockhttpf.checkConnectionArg2, Equals, 9050)

	i := ix.(*instance)
	c.Assert(i.started, Equals, true)
	c.Assert(i.socksPort, Equals, 9050)
	c.Assert(i.controlHost, Equals, "127.0.0.1")
	c.Assert(i.controlPort, Equals, 9051)
	c.Assert(i.useCookie, Equals, false)
	c.Assert(i.isLocal, Equals, true)
	c.Assert(i.runningTor, IsNil)
	c.Assert(i.binary, IsNil)
}

func (s *TorAcceptanceSuite) Test_thatSystemTorIsUsed_whenSystemTorIsOKWithCookieAuthAndProperVersion(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	tc := &mockTorgoController{}
	tc.authNoneReturn = errors.New("couldn't authenticate")
	tc.authPassReturn = errors.New("couldn't auth")
	tc.authCookieReturn = nil
	tc.getVersionReturn1 = "4.0.2"
	tc.getVersionReturn2 = nil

	mocktorgof.newControllerReturn1 = tc

	mockhttpf.checkConnectionReturn = true

	ix, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, IsNil)

	c.Assert(mocktorgof.newControllerArg, Equals, "127.0.0.1:9051")

	c.Assert(tc.authNoneCalled, Equals, 1)
	c.Assert(tc.authPassCalled, Equals, 0)
	c.Assert(tc.authCookieCalled, Equals, 2)

	c.Assert(tc.getVersionCalled, Equals, 1)

	c.Assert(mockhttpf.checkConnectionArg1, Equals, "127.0.0.1")
	c.Assert(mockhttpf.checkConnectionArg2, Equals, 9050)

	i := ix.(*instance)
	c.Assert(i.started, Equals, true)
	c.Assert(i.socksPort, Equals, 9050)
	c.Assert(i.controlHost, Equals, "127.0.0.1")
	c.Assert(i.controlPort, Equals, 9051)

	c.Assert(i.useCookie, Equals, true)
	c.Assert(i.isLocal, Equals, true)
	c.Assert(i.runningTor, IsNil)
	c.Assert(i.binary, IsNil)
}

func (s *TorAcceptanceSuite) Test_thatSystemTorWillBeUsed_whenSystemTorIsOKWithPasswordAuthAndOKVersion(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	*config.TorControlPassword = "super secret samosa"
	defer func() {
		*config.TorControlPassword = ""
	}()

	tc := &mockTorgoController{}
	tc.authNoneReturn = errors.New("couldn't authenticate")
	tc.authPassReturn = nil
	tc.authCookieReturn = errors.New("couldn't authenticate")
	tc.getVersionReturn1 = "4.0.3"
	tc.getVersionReturn2 = nil

	mocktorgof.newControllerReturn1 = tc

	mockhttpf.checkConnectionReturn = true

	ix, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, IsNil)

	c.Assert(mocktorgof.newControllerArg, Equals, "127.0.0.1:9051")

	c.Assert(tc.authNoneCalled, Equals, 1)
	c.Assert(tc.authPassCalled, Equals, 2)
	c.Assert(tc.authCookieCalled, Equals, 1)
	c.Assert(tc.authPassArg, Equals, "super secret samosa")

	c.Assert(tc.getVersionCalled, Equals, 1)

	c.Assert(mockhttpf.checkConnectionArg1, Equals, "127.0.0.1")
	c.Assert(mockhttpf.checkConnectionArg2, Equals, 9050)

	i := ix.(*instance)
	c.Assert(i.started, Equals, true)
	c.Assert(i.socksPort, Equals, 9050)
	c.Assert(i.controlHost, Equals, "127.0.0.1")
	c.Assert(i.controlPort, Equals, 9051)

	c.Assert(i.useCookie, Equals, false)
	c.Assert(i.isLocal, Equals, true)
	c.Assert(i.runningTor, IsNil)
	c.Assert(i.binary, IsNil)
	c.Assert(i.password, Equals, "super secret samosa")
}

func (s *TorAcceptanceSuite) Test_thatSystemTorWillNotBeShutDown_whenSystemTorIsUsed(c *C) {
	// TODO: figure out later
}

func (s *TorAcceptanceSuite) Test_thatSystemTorWillNotBeUsed_whenItsNotConnectedToTheInternet(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	tc := &mockTorgoController{}
	tc.authNoneReturn = nil
	tc.authPassReturn = errors.New("couldn't auth")
	tc.authCookieReturn = errors.New("couldn't auth")
	tc.getVersionReturn1 = "4.0.4"
	tc.getVersionReturn2 = nil

	mocktorgof.newControllerReturn1 = tc

	mockhttpf.checkConnectionReturn = false

	_, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, ErrorMatches, "no Tor binary found")
}

func (s *TorAcceptanceSuite) Test_thatSystemTorWillNotBeUsed_whenTheVersionIsTooOld(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	tc := &mockTorgoController{}
	tc.authNoneReturn = nil
	tc.authPassReturn = errors.New("couldn't auth")
	tc.authCookieReturn = errors.New("couldn't auth")
	tc.getVersionReturn1 = "0.2.1.1"
	tc.getVersionReturn2 = nil

	mocktorgof.newControllerReturn1 = tc

	mockhttpf.checkConnectionReturn = true

	_, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, ErrorMatches, "no Tor binary found")
}

func (s *TorAcceptanceSuite) Test_thatThingsWillFailIfTheresNoSystemTor(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	mocktorgof.newControllerReturn2 = errors.New("no connection possible")

	_, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, ErrorMatches, "no Tor binary found")
}

func (s *TorAcceptanceSuite) Test_thatThingsWillFailIfTheresASystemTorWithOldVersion(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	mocktorgof.newControllerReturn2 = errors.New("no connection possible")

	mockexecf.lookPathReturn1 = systemTorBinary

	calledAfter := 0
	called := false
	mockexecf.onExecWithModify = func(s string, a []string, mm ModifyCommand) ([]byte, error) {
		if called {
			calledAfter++
		}
		if s == systemTorBinary && len(a) > 0 && a[0] == "--version" && !called {
			called = true
			return []byte("Tor version 0.2.2.6."), nil
		}
		return nil, nil
	}

	_, e := InitializeInstance(&config.ApplicationConfig{})

	c.Assert(e, ErrorMatches, "no Tor binary found")
	c.Assert(called, Equals, true)
	c.Assert(calledAfter, Equals, 0)
}

func (s *TorAcceptanceSuite) Test_thatASystemTorBinaryWillBeStartedIfProperVersion(c *C) {
	mockAll()
	defer setDefaultFacades()
	defer func() {
		currentInstance = nil
	}()
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(ioutil.Discard)

	tc := &mockTorgoController{}
	tc.authNoneReturn = errors.New("couldn't authenticate")
	tc.authPassReturn = errors.New("couldn't auth")
	tc.authCookieReturn = nil
	tc.getVersionReturn1 = "4.0.2"
	tc.getVersionReturn2 = nil

	mocktorgof.onNewController = func(a string) (torgoController, error) {
		if a == "127.0.0.1:4215" {
			return tc, nil
		}
		return nil, errors.New("no connection possible")
	}

	mockhttpf.checkConnectionReturn = true

	setupSomeBasicMocks()

	finishWaiting := make(chan bool)
	mockexecf.onWaitCommand = func(c *exec.Cmd) error {
		<-finishWaiting
		return nil
	}

	ix, e := InitializeInstance(&config.ApplicationConfig{})

	verifyAllAssertions(c, e, tc, ix.(*instance))

	finishWaiting <- true
}

func setupSomeBasicMocks() {
	mockexecf.lookPathReturn1 = systemTorBinary

	mockexecf.onExecWithModify = func(s string, a []string, mm ModifyCommand) ([]byte, error) {
		if s == systemTorBinary && len(a) > 0 && a[0] == "--version" {
			return []byte("Tor version 0.4.2.1."), nil
		}
		return nil, nil
	}

	mockfilesystemf.onTempDir = func(s, s2 string) (string, error) {
		if s == config.WithHome(".local/share/wahay") {
			return config.WithHome(".local/share/wahay/4215-tor"), nil
		}
		return "", nil
	}

	mockosf.onIsPortAvailable = func(p int) bool {
		return p == 4215 || p == 4666
	}

	portCalled := 0
	mockosf.onGetRandomPort = func() int {
		portCalled++
		switch portCalled {
		case 1:
			return 4215
		case 2:
			return 12234
		case 3:
			return 4666
		default:
			return 0
		}
	}
}

func verifyAllAssertions(c *C, e error, tc *mockTorgoController, i *instance) {
	c.Assert(e, IsNil)

	c.Assert(tc.authNoneCalled, Equals, 1)
	c.Assert(tc.authPassCalled, Equals, 0)
	c.Assert(tc.authCookieCalled, Equals, 2)

	c.Assert(tc.getVersionCalled, Equals, 1)

	c.Assert(mockhttpf.checkConnectionArg1, Equals, "127.0.0.1")
	c.Assert(mockhttpf.checkConnectionArg2, Equals, 4666)

	c.Assert(i.started, Equals, true)
	c.Assert(i.socksPort, Equals, 4666)
	c.Assert(i.controlHost, Equals, "127.0.0.1")
	c.Assert(i.controlPort, Equals, 4215)

	c.Assert(i.useCookie, Equals, true)
	c.Assert(i.isLocal, Equals, false)
	c.Assert(i.password, Equals, "")
	c.Assert(i.configFile, Equals, config.WithHome(".local/share/wahay/4215-tor/torrc"))
	c.Assert(i.dataDirectory, Equals, config.WithHome(".local/share/wahay/4215-tor/data"))

	c.Assert(i.runningTor.cmd.Path, Equals, systemTorBinary)
	c.Assert(i.runningTor.cmd.Args, DeepEquals, []string{systemTorBinary, "-f",
		config.WithHome(".local/share/wahay/4215-tor/torrc")})
	c.Assert(i.runningTor.cmd.Env, IsNil)
	c.Assert(i.runningTor.cmd.Dir, Equals, "")
	c.Assert(i.runningTor.finished, Equals, false)
	c.Assert(i.runningTor.finishedWithError, IsNil)

	c.Assert(i.binary.path, Equals, systemTorBinary)
	c.Assert(i.binary.env, DeepEquals, []string{})
	c.Assert(i.binary.isValid, Equals, true)
	c.Assert(i.binary.isBundle, Equals, false)
}

// WHEN system tor instance can't be used:
// ---------------------------------------
// - there is no tor found anywhere
// - there is a "bundle" tor available
// - there is a system tor executable available

const shouldTestPrint = false

func testPrint(s string, args ...interface{}) {
	if shouldTestPrint {
		fmt.Printf(s, args...)
	}
}

var mockosf *mockOsImplementation
var mockfilepathf *mockFilepathImplementation
var mockexecf *mockExecImplementation
var mockfilesystemf *mockFilesystemImplementation
var mocktorgof *mockTorgoImplementation
var mockhttpf *mockHTTPImplementation

func mockAll() {
	mockosf = &mockOsImplementation{}
	mockfilepathf = &mockFilepathImplementation{}
	mockexecf = &mockExecImplementation{}
	mockfilesystemf = &mockFilesystemImplementation{}
	mocktorgof = &mockTorgoImplementation{}
	mockhttpf = &mockHTTPImplementation{}

	osf = mockosf
	filepathf = mockfilepathf
	execf = mockexecf
	filesystemf = mockfilesystemf
	torgof = mocktorgof
	httpf = mockhttpf
}

type mockOsImplementation struct {
	onIsPortAvailable func(int) bool
	onGetRandomPort   func() int
}

func (*mockOsImplementation) Getwd() (string, error) {
	testPrint("Getwd()\n")
	return "", nil
}

func (*mockOsImplementation) Args() []string {
	testPrint("Args()\n")
	return []string{"wahayTest"}
}

func (*mockOsImplementation) Environ() []string {
	testPrint("Environ()\n")
	return nil
}

func (*mockOsImplementation) RemoveAll(dir string) error {
	testPrint("RemoveAll(%s)\n", dir)
	return nil
}

func (*mockOsImplementation) MkdirAll(dir string, mode os.FileMode) error {
	testPrint("MkdirAll(%s, %v)\n", dir, mode)
	return nil
}

func (*mockOsImplementation) Stdout() *os.File {
	testPrint("Stdout()\n")
	return nil
}

func (*mockOsImplementation) Stderr() *os.File {
	testPrint("Stderr()\n")
	return nil
}

func (m *mockOsImplementation) IsPortAvailable(port int) bool {
	testPrint("IsPortAvailable(%v)\n", port)
	if m.onIsPortAvailable != nil {
		return m.onIsPortAvailable(port)
	}
	return true
}

func (m *mockOsImplementation) GetRandomPort() int {
	testPrint("GetRandomPort()\n")
	if m.onGetRandomPort != nil {
		return m.onGetRandomPort()
	}
	return 0
}

type mockFilepathImplementation struct{}

func (*mockFilepathImplementation) Glob(p string) ([]string, error) {
	testPrint("Glob(%v)\n", p)
	return nil, nil
}

type mockExecImplementation struct {
	lookPathReturn1 string
	lookPathReturn2 error

	onExecWithModify func(string, []string, ModifyCommand) ([]byte, error)
	onStartCommand   func(cmd *exec.Cmd) error
	onWaitCommand    func(cmd *exec.Cmd) error
}

func (m *mockExecImplementation) LookPath(s string) (string, error) {
	testPrint("LookPath(%v)\n", s)
	return m.lookPathReturn1, m.lookPathReturn2
}

func (m *mockExecImplementation) ExecWithModify(bin string, args []string, cm ModifyCommand) ([]byte, error) {
	testPrint("ExecWithModify(%v, %v, %v)\n", bin, args, cm)
	if m.onExecWithModify != nil {
		return m.onExecWithModify(bin, args, cm)
	}
	return nil, nil
}

func (m *mockExecImplementation) StartCommand(cmd *exec.Cmd) error {
	testPrint("StartCommand(%v)\n", cmd)
	if m.onStartCommand != nil {
		return m.onStartCommand(cmd)
	}
	return nil
}

func (m *mockExecImplementation) WaitCommand(cmd *exec.Cmd) error {
	testPrint("WaitCommand(%v)\n", cmd)
	if m.onWaitCommand != nil {
		return m.onWaitCommand(cmd)
	}
	return nil
}

type mockFilesystemImplementation struct {
	onTempDir   func(string, string) (string, error)
	onEnsureDir func(string, os.FileMode)
	onWriteFile func(string, []byte, os.FileMode) error
}

func (*mockFilesystemImplementation) FileExists(path string) bool {
	testPrint("FileExists(%v)\n", path)
	return false
}

func (*mockFilesystemImplementation) IsADirectory(path string) bool {
	testPrint("IsADirectory(%v)\n", path)
	return false
}

func (m *mockFilesystemImplementation) TempDir(where, suffix string) (string, error) {
	testPrint("TempDir(%v, %v)\n", where, suffix)
	if m.onTempDir != nil {
		return m.onTempDir(where, suffix)
	}
	return "", nil
}

func (m *mockFilesystemImplementation) EnsureDir(name string, mode os.FileMode) {
	testPrint("EnsureDir(%v, %v)\n", name, mode)
	if m.onEnsureDir != nil {
		m.onEnsureDir(name, mode)
	}
}

func (m *mockFilesystemImplementation) WriteFile(name string, content []byte, mode os.FileMode) error {
	testPrint("WriteFile(%v, %v, %v)\n", name, content, mode)
	if m.onWriteFile != nil {
		return m.onWriteFile(name, content, mode)
	}
	return nil
}

type mockTorgoController struct {
	authNoneReturn, authPassReturn, authCookieReturn error
	authNoneCalled, authPassCalled, authCookieCalled int

	authPassArg string

	getVersionReturn1 string
	getVersionReturn2 error
	getVersionCalled  int
}

func (m *mockTorgoController) AuthenticatePassword(v string) error {
	testPrint("torgoController.AuthenticatePassword(%v)\n", v)
	m.authPassCalled++
	m.authPassArg = v
	return m.authPassReturn
}

func (m *mockTorgoController) AuthenticateCookie() error {
	testPrint("torgoController.AuthenticateCookie()\n")
	m.authCookieCalled++
	return m.authCookieReturn
}

func (m *mockTorgoController) AuthenticateNone() error {
	testPrint("torgoController.AuthenticateNone()\n")
	m.authNoneCalled++
	return m.authNoneReturn
}

func (m *mockTorgoController) AddOnion(o *torgo.Onion) error {
	testPrint("torgoController.AddOnion(%v)\n", o)
	return nil
}

func (m *mockTorgoController) GetVersion() (string, error) {
	testPrint("torgoController.GetVersion()\n")
	m.getVersionCalled++
	return m.getVersionReturn1, m.getVersionReturn2
}

func (m *mockTorgoController) DeleteOnion(v string) error {
	testPrint("torgoController.DeleteOnion(%v)\n", v)
	return nil
}

type mockTorgoImplementation struct {
	newControllerArg     string
	newControllerReturn1 torgoController
	newControllerReturn2 error

	onNewController func(a string) (torgoController, error)
}

func (m *mockTorgoImplementation) NewController(a string) (torgoController, error) {
	testPrint("NewController(%v)\n", a)
	if m.onNewController != nil {
		return m.onNewController(a)
	}
	m.newControllerArg = a
	return m.newControllerReturn1, m.newControllerReturn2
}

type mockHTTPImplementation struct {
	checkConnectionArg1   string
	checkConnectionArg2   int
	checkConnectionReturn bool
}

func (m *mockHTTPImplementation) CheckConnectionOverTor(host string, port int) bool {
	testPrint("CheckConnectionOverTor(%v, %v)\n", host, port)
	m.checkConnectionArg1 = host
	m.checkConnectionArg2 = port
	return m.checkConnectionReturn
}

func (m *mockHTTPImplementation) HTTPRequest(host string, port int, u string) (string, error) {
	testPrint("HTTPRequest(%v, %v, %v)\n", host, port, u)
	return "", nil
}
