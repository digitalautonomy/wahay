package hosting

import (
	"errors"

	grumbleServer "github.com/digitalautonomy/grumble/server"
	"github.com/prashantv/gostub"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func (s *hostingSuite) Test_GenerateURL_returnsEmptyMumbleURLWhenNoMeetingDataHasBeenGiven(c *C) {
	md := MeetingData{}
	url := md.GenerateURL()
	c.Assert(url, Equals, "mumble://:@:0")
}

func (s *hostingSuite) Test_GenerateURL_returnsValidMumbleURLWhenAllMeetingDataHasBeenGiven(c *C) {
	md := MeetingData{
		MeetingID: "meetingId",
		Port:      23840,
		Password:  "mypassword",
		Username:  "TestUser",
	}

	url := md.GenerateURL()
	c.Assert(url, Equals, "mumble://TestUser:mypassword@meetingId:23840")
}

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
	expectedDataDir := `/tmp/wahay\d+$`

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
	mtd.On("TempDir", "", "wahay").Return("", errors.New("unknown error"))

	err := servers.initializeDataDirectory()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "unknown error")
	mtd.AssertExpectations(c)
}

func (s *hostingSuite) Test_callAll_executesAllIntroducedFunctions(c *C) {
	err := callAll(
		noErrHelper,
		errHelper)

	c.Assert(err.Error(), Equals, "error2")
}

func (s *hostingSuite) Test_startListener_setTrueIntoServersStartedStatus(c *C) {
	servers := &servers{}
	servers.startListener()
	c.Assert(servers.started, Equals, true)
}

func (s *hostingSuite) Test_initializeCertificates_emptyServersInstanceReturnsNotSuchFileOrDirectoryError(c *C) {
	servers := &servers{}
	servers.log = log.New() //Must have a log or panics
	expectedErr := `open .*/cert.pem: no such file or directory`

	err := servers.initializeCertificates()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Matches, expectedErr)
}

//TODO: These tests are working, but they're generating files that we don't need in testing mode
//I have to fix this with some mocks but that is not going to be today

// func (s *hostingSuite) Test_initializeLogging_emptyServersInstanceReturnsNoError(c *C) {
// 	servers := &servers{}
// 	err := servers.initializeLogging()
// 	c.Assert(err, IsNil)
// }

// func (s *hostingSuite) Test_initializeLogging_emptyServersInstanceReturnsNoError(c *C) {
// 	servers := &servers{}
// 	err := servers.initializeLogging()
// 	c.Assert(err, IsNil)
// }

// func (s *hostingSuite) Test_initializeLogging_verifyThatServerLogHasBeenCreated(c *C) {
// 	servers := &servers{}

// 	c.Assert(servers.log, IsNil)
// 	servers.initializeLogging()
// 	c.Assert(servers.log, NotNil)
// }

// func (s *hostingSuite) Test_servers_create_emptyServersInstanceReturnsNoError(c *C) {
// 	servers := &servers{}
// 	err := servers.create()
// 	c.Assert(err, IsNil)
// }

// func (s *hostingSuite) Test_servers_create_callFunctionTwiceShouldReturnAnErrorButItDoesnt(c *C) {
// 	servers := &servers{}

// 	servers.create()
// 	err := servers.create()
// 	c.Assert(err, IsNil)
// 	//This scenario should return an advice, an error or something
// }

// func (s *hostingSuite) Test_create_createServerCollection(c *C) {
// 	servers, err := create()

// 	c.Assert(servers, NotNil)
// 	c.Assert(err, IsNil)
// }
