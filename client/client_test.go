package client

import (
	"errors"

	. "github.com/digitalautonomy/wahay/test"
	"github.com/digitalautonomy/wahay/tor"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

type mockTempDir struct {
	mock.Mock
}

func (m *mockTempDir) tempDir(dir, prefix string) (string, error) {
	args := m.Called(dir, prefix)
	return args.String(0), args.Error(1)
}

func (s *clientSuite) Test_tempFolder_createsATempFolderSuccessfully(c *C) {
	tempDir, err := tempFolder()

	c.Assert(tempDir, NotNil)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_tempFolder_returnsAnErrorWhenTheTempFolderIsNotCreated(c *C) {

	mtd := &mockTempDir{}

	defer gostub.New().Stub(&tempDir, mtd.tempDir).Reset()

	mtd.On("tempDir", "", "mumble").Return("", errors.New("Error creating the temp folder")).Once()
	tempDir, err := tempFolder()

	c.Assert(tempDir, Equals, "")
	c.Assert(err, NotNil)

	mtd.AssertExpectations(c)
}

func (s *clientSuite) Test_LastError_returnsClientError(c *C) {
	client := &client{err: errors.New("Client Error")}

	result := client.LastError()

	c.Assert(result, NotNil)
}

func (s *clientSuite) Test_pathToBinary_returnsTheValidBinaryPath(c *C) {
	client := &client{isValid: true, binary: &binary{path: "path/to/binary"}}

	result := client.pathToBinary()

	c.Assert(result, Equals, "path/to/binary")
}

func (s *clientSuite) Test_pathToBinary_returnsAnEmptyStringWhenTheClientIsNotValid(c *C) {
	client := &client{isValid: false}

	result := client.pathToBinary()

	c.Assert(result, Equals, "")
}

func (s *clientSuite) Test_IsValid_returnsTrueWhenTheClientIsValid(c *C) {
	client := &client{isValid: true}

	result := client.IsValid()

	c.Assert(result, IsTrue)
}

func (s *clientSuite) Test_IsValid_returnsFalseWhenTheClientIsNotValid(c *C) {
	client := &client{isValid: false}

	result := client.IsValid()

	c.Assert(result, IsFalse)
}

func (s *clientSuite) Test_validate_returnsAnErrorWhenTheClientBinaryIsNilAndSetsIsValidToFalse(c *C) {
	client := &client{binary: nil}

	result := client.validate()
	c.Assert(result, Equals, errInvalidBinary)
	c.Assert(client.isValid, IsFalse)
}

func (s *clientSuite) Test_validate_returnsAnErrorWhenTheClientBinaryIsInvalidAndSetsIsValidToFalse(c *C) {
	client := &client{binary: &binary{isValid: false}}

	result := client.validate()
	c.Assert(result, Equals, errInvalidBinary)
	c.Assert(client.isValid, IsFalse)
}

func (s *clientSuite) Test_validate_returnsNilAndSetsIsValidToTrueWhenTheBinaryIsValid(c *C) {
	client := &client{binary: &binary{isValid: true}}

	result := client.validate()
	c.Assert(result, IsNil)
	c.Assert(client.isValid, IsTrue)
	c.Assert(client.err, IsNil)
}

func (s *clientSuite) Test_setBinary_returnsNilAndsetsClientBinaryIfTheBinaryIsValid(c *C) {
	binary := &binary{isValid: true}

	client := &client{}

	err := client.setBinary(binary)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_setBinary_returnsAnErrorIfTheBinaryIsNotValid(c *C) {
	binary := &binary{isValid: false}

	client := &client{}

	err := client.setBinary(binary)
	c.Assert(err, ErrorMatches, "the provided binary is not valid")
}

func (s *clientSuite) Test_torCommandModifier_returnsNilWhenTheClientIsNotValid(c *C) {
	client := &client{isValid: false}

	result := client.torCommandModifier()
	c.Assert(result, IsNil)
}

func (s *clientSuite) Test_torCommandModifier_setsClientTorCmdModifierAndReturnsATorModifyCommandWhenTheClientIsValid(c *C) {
	client := &client{isValid: true}
	c.Assert(client.torCmdModifier, IsNil)

	var expected tor.ModifyCommand
	result := client.torCommandModifier()
	c.Assert(client.torCmdModifier, NotNil)
	c.Assert(result, FitsTypeOf, expected)
}

func (s *clientSuite) Test_binaryEnv_appendsEnvironmentVariableWhenTheClientIsValidAndTheBinaryIsNotNil(c *C) {
	client := &client{binary: &binary{env: []string{"ENVIRONMENT=variable"}, isBundle: true}, isValid: true}
	envVariables := client.binaryEnv()

	c.Assert(envVariables, DeepEquals, []string{"QT_QPA_PLATFORM=xcb", "ENVIRONMENT=variable"})
}

func (s *clientSuite) Test_binaryEnv_returnsEnvironmentVariableWhenClientIsNotValidAndHasNoBinary(c *C) {
	client := &client{binary: nil, isValid: false}
	envVariable := client.binaryEnv()

	c.Assert(envVariable, DeepEquals, []string{"QT_QPA_PLATFORM=xcb"})
}
