package hosting

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/digitalautonomy/grumble/pkg/logtarget"
	grumbleServer "github.com/digitalautonomy/grumble/server"
	"github.com/digitalautonomy/wahay/tor"
	"github.com/prashantv/gostub"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

type mockStat struct {
	mock.Mock
}

func (m *mockStat) Stat(name string) (fs.FileInfo, error) {
	ret := m.Called(name)
	return nil, ret.Error(1)
}

func (h *hostingSuite) Test_defaultHost_returnsLocalhostInterfaceWhenWorkstationFileHasNotBeenFound(c *C) {
	ms := &mockStat{}

	defer gostub.New().Stub(&stat, ms.Stat).Reset()
	ms.On("Stat", "/usr/share/anon-ws-base-files/workstation").Return(nil, fs.ErrNotExist).Once()

	dh := defaultHost()
	localhostInterface := "127.0.0.1"

	c.Assert(dh, Equals, localhostInterface)
	ms.AssertExpectations(c)
}

func (h *hostingSuite) Test_defaultHost_returnsAllInterfacesWhenWorkstationFileHasBeenFound(c *C) {
	ms := &mockStat{}

	defer gostub.New().Stub(&stat, ms.Stat).Reset()
	ms.On("Stat", "/usr/share/anon-ws-base-files/workstation").Return(nil, nil).Once()

	dh := defaultHost()
	allInterfaces := "0.0.0.0"
	c.Assert(dh, Equals, allInterfaces)
	ms.AssertExpectations(c)
}

func (h *hostingSuite) Test_defaultHost_returnsLocalhostInterfaceWhenSomeKindOfErrorOcurred(c *C) {
	hook := logtest.NewGlobal()
	defer hook.Reset()
	log.SetOutput(io.Discard)

	ms := &mockStat{}

	defer gostub.New().Stub(&stat, ms.Stat).Reset()
	ms.On("Stat", "/usr/share/anon-ws-base-files/workstation").Return(nil, errors.New("unknown error related to Stat")).Once()

	dh := defaultHost()
	localhostInterface := "127.0.0.1"

	c.Assert(dh, Equals, localhostInterface)
	ms.AssertExpectations(c)
}

func (h *hostingSuite) Test_SetWelcomeText_worksWithBasicExample(c *C) {
	srvc := &service{}
	message := "This is a Wahay service test"
	srvc.SetWelcomeText(message)
	c.Assert(srvc.welcomeText, Equals, message)
}

func (h *hostingSuite) Test_NewService_returnsAnErrorWhenFailsCreatingCertificateServerBecauseNoDataDirectoryExists(c *C) {
	servers := &servers{}
	var ti tor.Instance
	srvc, err := servers.NewService("", ti)

	expectedErr := "the certificate file do not exists"

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, expectedErr)
	c.Assert(srvc, IsNil)
}

func (h *hostingSuite) Test_NewService_returnsAnErrorWhenWrongPortIsGiven(c *C) {
	path := "/tmp/wahay"
	var perm fs.FileMode = 0700

	e := os.MkdirAll(path, perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}
	defer os.RemoveAll(path)

	file := "cert.pem"
	fp := filepath.Join(path, file)
	_, e = os.Create(fp)
	if e != nil {
		c.Fatalf("Failed to create file: %v", e)
	}

	servers := &servers{
		dataDir: path,
	}
	var ti tor.Instance
	srvc, err := servers.NewService("xx", ti)

	c.Assert(err, NotNil)
	c.Assert(err, Equals, errInvalidPort)
	c.Assert(srvc, IsNil)
}

func (s *hostingSuite) Test_NewConferenceRoom_returnsAnErrorWhenFailsCreatingServer(c *C) {
	path := "/tmp/wahay/"
	sID := 2
	servers := &servers{
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
		nextID:  sID,
	}
	srvc := &service{
		collection: servers,
	}
	sud := SuperUserData{}
	err := srvc.NewConferenceRoom("", sud)
	expectedError := "mkdir " + path + "servers/" + strconv.Itoa(sID+1) + ": no such file or directory"
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, expectedError)
}

func (s *hostingSuite) Test_NewConferenceRoom_returnsAnErrorWhenFailsStartingServer(c *C) {
	path := "/tmp/wahay/"
	var perm fs.FileMode = 0700

	e := os.MkdirAll(filepath.Join(path, "servers"), perm)
	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}
	defer os.RemoveAll(path)

	servers := &servers{
		servers: make(map[int64]*grumbleServer.Server),
		dataDir: path,
		nextID:  1,
	}
	grumbleServer.Args.DataDir = path
	srvc := &service{
		collection: servers,
	}
	sud := SuperUserData{}
	err := srvc.NewConferenceRoom("", sud)
	expectedError := `open .*/cert.pem: no such file or directory`
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Matches, expectedError)
}

func (s *hostingSuite) Test_NewConferenceRoom_returnsNilWhenSuccessfullyCreatesANewConferenceRoom(c *C) {
	servers := &servers{
		nextID: 2,
	}

	servers.initializeSharedObjects()
	servers.initializeDataDirectory()

	logDir := path.Join(servers.dataDir, "grumble.log")
	grumbleServer.Args.LogPath = logDir
	logtarget.Target.OpenFile(logDir)

	l := log.New()
	l.SetOutput(io.Discard)
	servers.log = l

	servers.initializeCertificates()
	srvc := &service{
		collection: servers,
		httpServer: &webserver{
			running: true,
			server:  &http.Server{},
			address: "127.0.0.1:5545",
		},
	}
	sud := SuperUserData{}
	err := srvc.NewConferenceRoom("", sud)

	c.Assert(err, IsNil)
}
