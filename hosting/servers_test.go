package hosting

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	grumbleServer "github.com/digitalautonomy/grumble/server"
	"github.com/prashantv/gostub"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func (s *hostingSuite) Test_initializeSharedObjects_checkIfServersMapHasBeenCreated(c *C) {
	servers := &servers{}
	sm := make(map[int64]*grumbleServer.Server)

	c.Assert(servers.servers, IsNil)

	servers.initializeSharedObjects()

	c.Assert(servers.servers, NotNil)
	c.Assert(servers.servers, DeepEquals, sm)
}

func (s *hostingSuite) Test_initializeDataDirectory_generateExpectedDataDirectory(c *C) {
	servers := &servers{}
	expectedDataDir := `.*[/\\]wahay\d+$`

	c.Assert(servers.dataDir, Equals, "")
	err := servers.initializeDataDirectory()
	c.Assert(servers.dataDir, Matches, expectedDataDir)
	c.Assert(err, IsNil)
}

type mockTempDir struct {
	mock.Mock
}

func (m *mockTempDir) TempDir(dir string, pattern string) (name string, err error) {
	ret := m.Called(dir, pattern)
	return ret.String(0), ret.Error(1)
}

func (s *hostingSuite) Test_initializeDataDirectory_returnsAnErrorWhenFailsCreatingWahayTemporalDirectory(c *C) {
	servers := &servers{}
	servers.log = log.New()
	mtd := &mockTempDir{}

	defer gostub.New().Stub(&ioutilTempDir, mtd.TempDir).Reset()
	mtd.On("TempDir", "", "wahay").Return("", errors.New("unknown error related to TempDir"))

	err := servers.initializeDataDirectory()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "unknown error related to TempDir")
	mtd.AssertExpectations(c)
}

type mockOs struct {
	mock.Mock
}

func (m *mockOs) MkdirAll(path string, perm fs.FileMode) error {
	ret := m.Called(path, perm)
	return ret.Error(0)
}

func (s *hostingSuite) Test_initializeDataDirectory_returnsAnErrorWhenFailsCreatingServersDirectory(c *C) {
	servers := &servers{}
	servers.log = log.New()

	mtd := &mockTempDir{}
	defer gostub.New().Stub(&ioutilTempDir, mtd.TempDir).Reset()
	mtd.On("TempDir", "", "wahay").Return("/tmp/wahay", nil)

	mo := &mockOs{}
	defer gostub.New().Stub(&osMkdirAll, mo.MkdirAll).Reset()
	var perm fs.FileMode = 0700
	mo.On("MkdirAll", filepath.Join("/tmp/wahay/servers"), perm).Return(errors.New("unknown error related to MkdirAll"))

	err := servers.initializeDataDirectory()
	c.Assert(err, ErrorMatches, "unknown error related to MkdirAll")

	mtd.AssertExpectations(c)
	mo.AssertExpectations(c)
}

func (s *hostingSuite) Test_initializeLogging_returnsAnErrorWhenIsNotPossibleToOpenGrumbleLogFile(c *C) {
	path := "/invalid/path/"
	servers := &servers{
		dataDir: path,
	}
	err := servers.initializeLogging()
	expectedError := "open " + path + "grumble.log: (no such file or directory|The system cannot find the path specified.)$"
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, expectedError)
}

func (s *hostingSuite) Test_initializeLogging_returnsNilWhenGrumbleLogHasBeenCreated(c *C) {
	hook := logtest.NewGlobal()
	defer hook.Reset()

	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700

	e := os.MkdirAll(path, perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}
	defer os.RemoveAll(path)

	servers := &servers{
		dataDir: path,
	}

	err := servers.initializeLogging()
	c.Assert(err, IsNil)
}

func (s *hostingSuite) Test_initializeCertificates_generatesSelfSignedCertificateWhenGrumbleDataDirIsCorrect(c *C) {
	servers := &servers{
		log: log.New(), //Must have a log or panics
	}

	var e error
	grumbleServer.Args.DataDir, e = os.MkdirTemp("", "wahay")

	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(grumbleServer.Args.DataDir)

	err := servers.initializeCertificates()
	c.Assert(err, IsNil)
}

func (s *hostingSuite) Test_initializeCertificates_returnsNotSuchFileOrDirectoryErrorWhenGrumbleDataDirIsNotSetted(c *C) {
	servers := &servers{
		log: log.New(),
	}
	expectedErr := `^open .*[/\\]cert.pem: (no such file or directory|The system cannot find the path specified.)$`

	err := servers.initializeCertificates()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, expectedErr)
}

func (s *hostingSuite) Test_callAll_executesAllIntroducedFunctions(c *C) {
	err := callAll(
		noErrHelper,
		errHelper)

	c.Assert(err.Error(), Equals, "error2")
}

func (s *hostingSuite) Test_startListener_setTrueIntoServersStartedStatus(c *C) {
	servers := &servers{}
	c.Assert(servers.started, Equals, false)
	servers.startListener()
	c.Assert(servers.started, Equals, true)
}

func (s *hostingSuite) Test_startListener_statusRemainsTheSameWhenServersIsAlreadyStarted(c *C) {
	servers := &servers{
		started: true,
	}
	c.Assert(servers.started, Equals, true)
	servers.startListener()
	c.Assert(servers.started, Equals, true)
}

func (s *hostingSuite) Test_CreateServer_setDefaultOptionsOnlyReturnsAServerInstance(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
	}

	serv, err := servers.CreateServer(setDefaultOptions)
	c.Assert(err, IsNil)
	c.Assert(reflect.TypeOf(serv), DeepEquals, reflect.TypeOf(&server{}))
}

func (s *hostingSuite) Test_CreateServer_setWelcomeTextOnlyReturnsAServerInstance(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
	}

	serv, err := servers.CreateServer(setWelcomeText("hello wahay"))
	c.Assert(err, IsNil)
	c.Assert(reflect.TypeOf(serv), DeepEquals, reflect.TypeOf(&server{}))
}

func (s *hostingSuite) Test_CreateServer_setPortOnlyReturnsAServerInstance(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
	}

	serv, err := servers.CreateServer(setPort("1234"))
	c.Assert(err, IsNil)
	c.Assert(reflect.TypeOf(serv), DeepEquals, reflect.TypeOf(&server{}))
}

func (s *hostingSuite) Test_CreateServer_setPasswordOnlyReturnsAServerInstance(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
	}

	serv, err := servers.CreateServer(setPassword("pAwd12!@"))
	c.Assert(err, IsNil)
	c.Assert(reflect.TypeOf(serv), DeepEquals, reflect.TypeOf(&server{}))
}

func (s *hostingSuite) Test_CreateServer_setSuperUserOnlyReturnsAServerInstance(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
	}

	serv, err := servers.CreateServer(setSuperUser("root", "pAwd12!@"))
	c.Assert(err, IsNil)
	c.Assert(reflect.TypeOf(serv), DeepEquals, reflect.TypeOf(&server{}))
}

func (s *hostingSuite) Test_CreateServer_sendSeveralServerModifiersReturnsAServerInstanceWithNoErrors(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
	}

	serv, err := servers.CreateServer(
		setDefaultOptions,
		setWelcomeText("hello wahay"),
		setPort("1234"),
		setPassword("pAwd12!@"),
		setSuperUser("root", "pAwd12!@"),
	)

	c.Assert(err, IsNil)
	c.Assert(reflect.TypeOf(serv), DeepEquals, reflect.TypeOf(&server{}))
}

func (s *hostingSuite) Test_CreateServer_returnsAnErrorWhenServersDirectoryHasNotBeenCreated(c *C) {
	servers := &servers{
		nextID:  1,
		servers: make(map[int64]*grumbleServer.Server),
	}

	expectedError := `mkdir servers[/\\]2: (no such file or directory|The system cannot find the path specified.)$`

	_, err := servers.CreateServer()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, expectedError)
}

func (s *hostingSuite) Test_DataDir_returnsTheCorrectDataDirectoryConfiguredOnTheStructure(c *C) {
	servers := &servers{
		dataDir: "example/dir",
	}

	dir := servers.DataDir()

	c.Assert(dir, Equals, "example/dir")
}

func (s *hostingSuite) Test_DataDir_returnsEmptyStringWhenDataDirectoryHasNotBeenConfigured(c *C) {
	servers := &servers{}

	dir := servers.DataDir()

	c.Assert(dir, Equals, "")
}

func (s *hostingSuite) Test_Cleanup_deleteCurrentServerDataDirSuccessfully(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(path, perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	expectedErr := "open " + path + ": (no such file or directory|The system cannot find the file specified.)$"

	servers := &servers{dataDir: path}
	_, e = os.ReadDir(path)
	c.Assert(e, IsNil)

	servers.Cleanup()

	_, e = os.ReadDir(path)
	c.Assert(e, NotNil)
	c.Assert(e, ErrorMatches, expectedErr)
}
