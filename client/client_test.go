package client

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/digitalautonomy/wahay/config"
	. "github.com/digitalautonomy/wahay/test"
	"github.com/digitalautonomy/wahay/tor"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
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

	if !IsWindows() {
		c.Assert(client.torCmdModifier, NotNil)
		c.Assert(result, FitsTypeOf, expected)
	} else {
		c.Assert(client.torCmdModifier, IsNil)
	}
}

func (s *clientSuite) Test_binaryEnv_appendsEnvironmentVariableWhenTheClientIsValidAndTheBinaryIsNotNil(c *C) {
	client := &client{binary: &binary{env: []string{"ENVIRONMENT=variable"}, isBundle: true}, isValid: true}
	envVariables := client.binaryEnv()

	if !IsWindows() {
		c.Assert(envVariables, DeepEquals, []string{"QT_QPA_PLATFORM=xcb", "ENVIRONMENT=variable"})
	} else {
		c.Assert(envVariables, DeepEquals, []string(nil))
	}
}

func (s *clientSuite) Test_binaryEnv_returnsEnvironmentVariableWhenClientIsNotValidAndHasNoBinary(c *C) {
	client := &client{binary: nil, isValid: false}
	envVariable := client.binaryEnv()

	if !IsWindows() {
		c.Assert(envVariable, DeepEquals, []string{"QT_QPA_PLATFORM=xcb"})
	} else {
		c.Assert(envVariable, DeepEquals, []string(nil))
	}
}

type MockTorInstance struct{}

func (m *MockTorInstance) Start() error {
	return nil
}

func (m *MockTorInstance) Destroy() {}

func (m *MockTorInstance) GetController() tor.Control {
	return nil
}

func (m *MockTorInstance) HTTPrequest(url string) (string, error) {
	return "mock response", nil
}

func (m *MockTorInstance) NewService(a string, b []string, c tor.ModifyCommand) (tor.Service, error) {
	return nil, nil
}

func (m *MockTorInstance) NewOnionServiceWithMultiplePorts(ports []tor.OnionPort) (tor.Onion, error) {
	return nil, nil
}

func (s *clientSuite) Test_InitSystem_worksWithAValidConfigurationAndBinaryPath(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	absolutePathForBinary := filepath.Join(tempDir, "path/to/binary")
	err = os.MkdirAll(absolutePathForBinary, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(absolutePathForBinary, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mc := &mockCommand{}
	mc.On("Command", srcf.Name(), []string{"--version"}).Return(&exec.Cmd{Path: srcf.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	ti := &MockTorInstance{}

	i := InitSystem(&config.ApplicationConfig{PathMumble: srcf.Name()}, ti)

	client := i.(*client)

	c.Assert(client.isValid, IsTrue)
	c.Assert(client.err, IsNil)

	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_InitSystem_returnsAnInvalidInstanceWhenAValidMumbleBinaryIsNotAvailable(c *C) {

	ml := &mockLookPath{}
	defer gostub.New().Stub(&execLookPath, ml.LookPath).Reset()
	ml.On("LookPath", "mumble").Return("", nil).Once()

	ti := &MockTorInstance{}

	i := InitSystem(&config.ApplicationConfig{}, ti)

	client := i.(*client)

	c.Assert(client.isValid, IsFalse)
	c.Assert(client.err, Equals, errBinaryUnavailable)

	ml.AssertExpectations(c)
}

func (s *clientSuite) Test_InitSystem_returnsAnInvalidInstanceWhenTemporaryFolderCreationFails(c *C) {
	tempDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDirectory)

	absolutePathForBinary := filepath.Join(tempDirectory, "path/to/binary")
	err = os.MkdirAll(absolutePathForBinary, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(absolutePathForBinary, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mc := &mockCommand{}
	mc.On("Command", srcf.Name(), []string{"--version"}).Return(&exec.Cmd{Path: srcf.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	ti := &MockTorInstance{}

	mtd := &mockTempDir{}

	defer gostub.New().Stub(&tempDir, mtd.tempDir).Reset()

	mtd.On("tempDir", "", "mumble").Return("", errors.New("Error creating the temp folder")).Once()

	i := InitSystem(&config.ApplicationConfig{PathMumble: srcf.Name()}, ti)

	client := i.(*client)

	c.Assert(client.isValid, IsFalse)
	c.Assert(client.err, ErrorMatches, "Error creating the temp folder")

	mc.AssertExpectations(c)
	mtd.AssertExpectations(c)
}

func (s *clientSuite) Test_InitSystem_returnsAnInvalidInstanceWhenEnsuringTheConfigurationFails(c *C) {
	tempDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDirectory)

	absolutePathForBinary := filepath.Join(tempDirectory, "path/to/binary")
	err = os.MkdirAll(absolutePathForBinary, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(absolutePathForBinary, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mc := &mockCommand{}
	mc.On("Command", srcf.Name(), []string{"--version"}).Return(&exec.Cmd{Path: srcf.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	ti := &MockTorInstance{}

	mm := &mockMkdirAll{}
	defer gostub.New().Stub(&osMkdirAll, mm.MkdirAll).Reset()
	var perm fs.FileMode = 0700
	mm.On("MkdirAll", mock.Anything, perm).Return(errors.New("Error creating directory")).Once()

	i := InitSystem(&config.ApplicationConfig{PathMumble: srcf.Name()}, ti)

	client := i.(*client)

	c.Assert(client.isValid, IsFalse)
	c.Assert(client.err, ErrorMatches, "invalid client configuration directory")

	mc.AssertExpectations(c)
}
