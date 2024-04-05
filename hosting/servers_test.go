package hosting

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

type ServersSuite struct{}

var _ = Suite(&ServersSuite{})

func Test(t *testing.T) { TestingT(t) }

func (s *ServersSuite) Test_GenerateURL_returnsEmptyMumbleURLWhenNoMeetingDataHasBeenGiven(c *C) {
	md := MeetingData{}
	url := md.GenerateURL()
	c.Assert(url, Equals, "mumble://:@:0")
}

func (s *ServersSuite) Test_GenerateURL_returnsValidMumbleURLWhenAllMeetingDataHasBeenGiven(c *C) {
	md := MeetingData{
		MeetingID: "meetingId",
		Port:      23840,
		Password:  "mypassword",
		Username:  "TestUser",
	}

	url := md.GenerateURL()
	c.Assert(url, Equals, "mumble://TestUser:mypassword@meetingId:23840")
}

func (s *ServersSuite) Test_initializeSharedObjects_checkIfServersMapHasBeenCreated(c *C) {
	servers := &servers{}

	c.Assert(servers.servers, IsNil)
	servers.initializeSharedObjects()
	c.Assert(servers.servers, NotNil)
}

func (s *ServersSuite) Test_initializeDataDirectory_generateExpectedDataDirectory(c *C) {
	servers := &servers{}
	expectedDataDir := `/tmp/wahay\d+$`

	c.Assert(servers.dataDir, Equals, "")
	err := servers.initializeDataDirectory()
	c.Assert(servers.dataDir, Matches, expectedDataDir)
	c.Assert(err, IsNil)
}

func (s *ServersSuite) Test_initializeLogging_emptyServersInstanceReturnsNoError(c *C) {
	servers := &servers{}
	err := servers.initializeLogging()
	c.Assert(err, IsNil)
}

func (s *ServersSuite) Test_initializeLogging_verifyThatServerLogHasBeenCreated(c *C) {
	servers := &servers{}

	c.Assert(servers.log, IsNil)
	servers.initializeLogging()
	c.Assert(servers.log, NotNil)
}

func (s *ServersSuite) Test_initializeCertificates_emptyServersInstanceReturnsNoError(c *C) {
	servers := &servers{}
	servers.log = log.New() //Must have a log or panics

	err := servers.initializeCertificates()
	c.Assert(err, IsNil)
}

func noErrHelper() error {
	return nil
}
func errHelper() error {
	return errors.New("error2")
}

func (s *ServersSuite) Test_callAll_executesAllIntroducedFunctions(c *C) {
	err := callAll(
		noErrHelper,
		errHelper)

	c.Assert(err.Error(), Equals, "error2")
}

func (s *ServersSuite) Test_servers_create_emptyServersInstanceReturnsNoError(c *C) {
	servers := &servers{}
	err := servers.create()
	c.Assert(err, IsNil)
}

func (s *ServersSuite) Test_servers_create_callFunctionTwiceReturnsAnError(c *C) {
	servers := &servers{}

	servers.create()
	err := servers.create()
	c.Assert(err, IsNil)
	//This scenario should return an advice, an error or something
}

func (s *ServersSuite) Test_create_createServerCollection(c *C) {
	servers, err := create()

	c.Assert(servers, NotNil)
	c.Assert(err, IsNil)
}
