package client

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/digitalautonomy/wahay/config"
	. "github.com/digitalautonomy/wahay/test"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_realBinaryPath_worksWithMumbleDirectoryPath(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	existingDir := filepath.Join(tempDir, "existing_dir")
	err = os.Mkdir(existingDir, 0755)
	if err != nil {
		c.Fatalf("Failed to create test directory: %v", err)
	}

	expected := filepath.Join(existingDir, mumbleBundlePath)
	result := realBinaryPath(existingDir)

	c.Assert(result, Equals, expected)
}

func (s *clientSuite) Test_realBinaryPath_worksWithMumbleBinaryPath(c *C) {

	result := realBinaryPath("/path/to/mumble/binary")

	c.Assert(result, Equals, "/path/to/mumble/binary")
}

func (s *clientSuite) Test_newBinary_worksWithValidBinaryPath(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumblePath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumblePath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binary := newBinary(tempDir)

	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.lastError, IsNil)
}

func (s *clientSuite) Test_newBinary_worksWithInvalidBinaryPath(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	binary := newBinary(tempDir)

	c.Assert(binary.isValid, IsFalse)
	c.Assert(binary.lastError, NotNil)
}

func (s *clientSuite) Test_checkLibsDependenciesInPath_worksWithMumbleBundledDependencies(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumblePath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumblePath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mumbleBundlePath := filepath.Join(tempDir, "/mumble/lib")
	err = os.MkdirAll(mumbleBundlePath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	isBundle, env := checkLibsDependenciesInPath(mumblePath)

	c.Assert(isBundle, IsTrue)
	c.Assert(env, HasLen, 1)
}

func (s *clientSuite) Test_checkLibsDependenciesInPath_worksWithMumbleWithoutBundledDependencies(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBundlePath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumbleBundlePath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	isBundle, env := checkLibsDependenciesInPath(mumbleBundlePath)

	c.Assert(isBundle, IsFalse)
	c.Assert(env, HasLen, 0)
}

func (s *clientSuite) Test_isThereAnAvailableBinary_worksWhenTheProvidedPathIsEmpty(c *C) {
	binary := isThereAnAvailableBinary("")
	c.Assert(binary.isValid, IsFalse)
}

type mockCommand struct {
	mock.Mock
}

func (m *mockCommand) Command(name string, arg ...string) *exec.Cmd {
	args := m.Called(name, arg)
	return args.Get(0).(*exec.Cmd)
}

func (m *mockCommand) Output() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (s *clientSuite) Test_isThereAnAvailableBinary_returnsAValidMumbleBundledBinaryWhenThePathIsCorrect(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleLibPath := filepath.Join(tempDir, "/mumble/lib")
	err = os.MkdirAll(mumbleLibPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mc := &mockCommand{}

	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()

	mc.On("Command", mumbleBinaryPath, []string{"--version"}).Return(&exec.Cmd{Path: mumbleBinaryPath, Args: []string{"--version"}}).Once()

	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary := isThereAnAvailableBinary(tempDir)

	c.Assert(binary.isBundle, IsTrue)
	c.Assert(binary.isValid, IsTrue)

	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_isThereAnAvailableBinary_returnsAInvalidMumbleBinaryWhenThePathIsIncorrect(c *C) {
	binary := isThereAnAvailableBinary("invalid/binary/path")

	c.Assert(binary.isValid, IsFalse)
}

func (s *clientSuite) Test_isThereAnAvailableBinary_returnsAValidAndNotBundledMumbleBinaryWhenThePathIsCorrect(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mc := &mockCommand{}

	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()

	mc.On("Command", mumbleBinaryPath, []string{"--version"}).Return(&exec.Cmd{Path: mumbleBinaryPath, Args: []string{"--version"}}).Once()

	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary := isThereAnAvailableBinary(tempDir)

	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.isBundle, IsFalse)

	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_isThereAnAvailableBinary_returnsAInValidMumbleBinaryOnInvalidCommand(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mc := &mockCommand{}

	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()

	mc.On("Output").Return([]byte{}, errors.New("Invalid Command")).Once()

	binary := isThereAnAvailableBinary(tempDir)

	c.Assert(binary.isValid, IsFalse)
	c.Assert(binary.lastError, Equals, errInvalidCommand)

	mc.AssertExpectations(c)
}

type mockLookPath struct {
	mock.Mock
}

func (m *mockLookPath) LookPath(file string) (string, error) {
	args := m.Called(file)
	return args.String(0), args.Error(1)
}

func (s *clientSuite) Test_searchBinaryInSystem_returnsAValidBinaryFoundInTheSystem(c *C) {
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
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	mc.On("Command", srcf.Name(), []string{"--version"}).Return(&exec.Cmd{Path: srcf.Name(), Args: []string{"--version"}}).Once()
	mc.On("Output").Return([]byte("command output"), nil).Once()

	ml := &mockLookPath{}
	defer gostub.New().Stub(&execLookPath, ml.LookPath).Reset()
	ml.On("LookPath", "mumble").Return(srcf.Name(), nil).Once()

	binary, err := searchBinaryInSystem()
	c.Assert(binary, NotNil)
	c.Assert(err, IsNil)
	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.lastError, IsNil)

	mc.AssertExpectations(c)
	ml.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInSystem_returnsNilWhenTheBinaryIsNotFoundInTheSystem(c *C) {
	ml := &mockLookPath{}
	defer gostub.New().Stub(&execLookPath, ml.LookPath).Reset()

	ml.On("LookPath", "mumble").Return("", errors.New("Error looking for executable")).Once()

	binary, _ := searchBinaryInSystem()
	c.Assert(binary, IsNil)

	ml.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInConf_returnedCallbackFunctionWorksWithAValidConfiguredPath(c *C) {
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
	conf := &config.ApplicationConfig{PathMumble: srcf.Name()}

	mc := &mockCommand{}
	mc.On("Command", srcf.Name(), []string{"--version"}).Return(&exec.Cmd{Path: srcf.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()

	callBack := searchBinaryInConf(conf)

	binary, err := callBack()

	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.lastError, IsNil)
	c.Assert(err, IsNil)

	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInConf_returnedCallbackFunctionReturnsNilWhenTheConfiguredPathIsEmpty(c *C) {
	conf := &config.ApplicationConfig{PathMumble: ""}

	callBack := searchBinaryInConf(conf)

	binary, err := callBack()

	c.Assert(binary, IsNil)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_searchBinaryInConf_returnedCallbackFunctionReturnsAnErrorWhenTheConfiguredPathIsInvalid(c *C) {
	conf := &config.ApplicationConfig{PathMumble: "invalid/binary/path"}

	callBack := searchBinaryInConf(conf)

	binary, err := callBack()

	c.Assert(binary, IsNil)
	c.Assert(err, Equals, errNoClientInConfiguredPath)
}

func (s *clientSuite) Test_searchBinaryInConf_returnsAValidFuncWhenANilConfIsProvided(c *C) {
	result := searchBinaryInConf(nil)
	c.Assert(result, NotNil)
}

func (s *clientSuite) Test_remove_removesTemporaryDirectoryContainingTheMumbleClient(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binary := &binary{isTemporary: true, path: mumbleBinaryPath}

	_, err = os.Stat(mumbleBinaryPath)

	c.Assert(err, IsNil)

	binary.remove()

	_, err = os.Stat(mumbleBinaryPath)

	c.Assert(err, NotNil)
}

func (s *clientSuite) Test_copyBinaryToDir_copyTheBinaryToANewFile(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mumbleBinaryDestinationPath := filepath.Join(tempDir, "/destination")
	err = os.MkdirAll(mumbleBinaryDestinationPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binary := &binary{path: srcf.Name()}

	// Assert that the Mumble binary does not exist in the destination directory
	_, err = os.Stat(mumbleBinaryDestinationPath + "/mumble")
	c.Assert(err, NotNil)

	err = binary.copyBinaryToDir(mumbleBinaryDestinationPath + "/mumble")
	c.Assert(err, IsNil)

	// Assert that the Mumble binary now exists in the destination directory
	_, err = os.Stat(mumbleBinaryDestinationPath + "/mumble")
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_copyBinaryToDir_returnsAnErrorWhenTheDestinationIsInvalid(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	binary := &binary{path: srcf.Name()}

	err = binary.copyBinaryToDir("invalid/binary/destination")
	c.Assert(err, NotNil)

	// Assert that the Mumble binary does not exist in the destination directory
	_, err = os.Stat("invalid/binary/destination")
	c.Assert(err, NotNil)
}

func (s *clientSuite) Test_copyBinaryToDir_returnsAnErrorWhenTheBinaryPathDoesNotExist(c *C) {

	binary := &binary{path: "invalid/binary/path"}

	err := binary.copyBinaryToDir("valid/binary/destination")
	c.Assert(err, NotNil)

	// Assert that the Mumble binary does not exist in the src directory
	_, err = os.Stat("invalid/binary/path")
	c.Assert(err, NotNil)
}

func (s *clientSuite) Test_copyBinaryToDir_returnsAnErrorWhenTheDestinationFileIsADirectory(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mumbleBinaryDestinationPath := filepath.Join(tempDir, "/destination")
	err = os.MkdirAll(mumbleBinaryDestinationPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binary := &binary{path: srcf.Name()}

	err = binary.copyBinaryToDir(mumbleBinaryDestinationPath)
	c.Assert(err, NotNil)

	c.Assert(isADirectory(mumbleBinaryDestinationPath), IsTrue)
}

func (s *clientSuite) Test_copyBinaryToDir_shouldOverwriteAnExistingFile(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mumbleBinaryDestinationPath := filepath.Join(tempDir, "/destination")
	err = os.MkdirAll(mumbleBinaryDestinationPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	dstf, err := os.CreateTemp(mumbleBinaryPath, "existingFile")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	binary := &binary{path: srcf.Name()}

	err = binary.copyBinaryToDir(dstf.Name())

	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_copyTo_worksWithAValidBinaryPathAndAValidDestinationPath(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	destinationBinaryPath := filepath.Join(tempDir, "/destination")
	err = os.MkdirAll(destinationBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binary := &binary{path: srcf.Name(), isValid: true}

	err = binary.copyTo(destinationBinaryPath)

	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_copyTo_returnsAnErrorIfTheBinaryPathDoesNotExist(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	binary := &binary{path: "invalid/binary/path", isValid: true}

	err = binary.copyTo("valid/binary/destination")

	c.Assert(err.Error(), Equals, errInvalidBinaryFile.Error())
}

func (s *clientSuite) Test_copyTo_returnsAnErrorIfDestinationPathIsNotADirectory(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	_, err = os.CreateTemp(tempDir, "destination")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	binary := &binary{path: srcf.Name(), isValid: true}

	err = binary.copyTo(tempDir + "/destination")

	c.Assert(err.Error(), Equals, errDestinationIsNotADirectory.Error())
}

type mockJoin struct {
	mock.Mock
}

func (m *mockJoin) Join(elem ...string) string {
	args := m.Called(elem)
	return args.String(0)
}

func (s *clientSuite) Test_copyTo_returnsAnErrorIfTheBinaryAlreadyExistInThePath(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	srcf, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	binaryDestinationPath := filepath.Join(tempDir, "/destination")
	err = os.MkdirAll(binaryDestinationPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	alreadyExistingBinary, err := os.CreateTemp(binaryDestinationPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	alreadyExistingBinaryPath := alreadyExistingBinary.Name()

	mj := &mockJoin{}
	defer gostub.New().Stub(&filepathJoin, mj.Join).Reset()
	mj.On("Join", []string{binaryDestinationPath, "mumble"}).Return(alreadyExistingBinaryPath).Once()

	binary := &binary{path: srcf.Name(), isValid: true}

	err = binary.copyTo(binaryDestinationPath)

	c.Assert(err.Error(), Equals, errBinaryAlreadyExists.Error())

	mj.AssertExpectations(c)
}

type mockGetwd struct {
	mock.Mock
}

func (m *mockGetwd) Getwd() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (s *clientSuite) Test_searchBinaryInCurrentWorkingDir_returnsAValidBinaryIfABinaryFileIsFoundInTheCurrentWorkingDirectory(c *C) {
	currentWorkingDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(currentWorkingDirectory)

	mumbleBinaryPath := filepath.Join(currentWorkingDirectory, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binaryFile, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mg := &mockGetwd{}
	defer gostub.New().Stub(&osGetwd, mg.Getwd).Reset()
	mg.On("Getwd").Return(currentWorkingDirectory, nil).Once()

	mj := &mockJoin{}
	defer gostub.New().Stub(&filepathJoin, mj.Join).Reset()
	mj.On("Join", []string{currentWorkingDirectory, mumbleBundlePath}).Return(binaryFile.Name())

	mc := &mockCommand{}
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	mc.On("Command", binaryFile.Name(), []string{"--version"}).Return(&exec.Cmd{Path: binaryFile.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary, err := searchBinaryInCurrentWorkingDir()
	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.lastError, IsNil)
	c.Assert(err, IsNil)

	mg.AssertExpectations(c)
	mj.AssertExpectations(c)
	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInCurrentWorkingDir_returnsNilIfThereIsAnErrorGettingTheCurrentDirectory(c *C) {
	currentWorkingDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(currentWorkingDirectory)

	mg := &mockGetwd{}
	defer gostub.New().Stub(&osGetwd, mg.Getwd).Reset()
	mg.On("Getwd").Return("", errors.New("Getwd error")).Once()

	binary, err := searchBinaryInCurrentWorkingDir()
	c.Assert(binary, IsNil)
	c.Assert(err, IsNil)

	mg.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInCurrentWorkingDir_returnsAInvalidBinaryIfABinaryFileIsNotFoundInTheCurrentWorkingDirectory(c *C) {
	currentWorkingDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(currentWorkingDirectory)

	mg := &mockGetwd{}
	defer gostub.New().Stub(&osGetwd, mg.Getwd).Reset()
	mg.On("Getwd").Return(currentWorkingDirectory, nil).Once()

	binary, err := searchBinaryInCurrentWorkingDir()
	c.Assert(binary.isValid, IsFalse)
	c.Assert(binary.lastError, ErrorMatches, "not valid binary path")
	c.Assert(err, IsNil)

	mg.AssertExpectations(c)
}

type mockAbs struct {
	mock.Mock
}

func (m *mockAbs) Abs(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

type mockDir struct {
	mock.Mock
}

func (m *mockDir) Dir(path string) string {
	args := m.Called(path)
	return args.String(0)
}

func (s *clientSuite) Test_searchBinaryInLocalDir_returnsAValidBinaryIfABinaryFileIsFoundInTheLocalDirectory(c *C) {
	localDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(localDirectory)

	mumbleBinaryPath := filepath.Join(localDirectory, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binaryFile, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mockOsArgs := []string{localDirectory + "/wahay"}
	osArgs = mockOsArgs

	ma := &mockAbs{}
	defer gostub.New().Stub(&filepathAbs, ma.Abs).Reset()
	ma.On("Abs", localDirectory).Return(localDirectory, nil).Once()

	md := &mockDir{}
	defer gostub.New().Stub(&filepathDir, md.Dir).Reset()
	md.On("Dir", osArgs[0]).Return(localDirectory).Once()

	mj := &mockJoin{}
	defer gostub.New().Stub(&filepathJoin, mj.Join).Reset()
	mj.On("Join", []string{localDirectory, mumbleBundlePath}).Return(binaryFile.Name())

	mc := &mockCommand{}
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	mc.On("Command", binaryFile.Name(), []string{"--version"}).Return(&exec.Cmd{Path: binaryFile.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary, err := searchBinaryInLocalDir()
	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.lastError, IsNil)
	c.Assert(err, IsNil)

	ma.AssertExpectations(c)
	md.AssertExpectations(c)
	mj.AssertExpectations(c)
	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInLocalDir_returnsNilWhenThereIsAnErrorGettingTheAbsolutePathOfTheRunningWahayProgram(c *C) {
	localDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(localDirectory)

	mumbleBinaryPath := filepath.Join(localDirectory, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mockOsArgs := []string{localDirectory + "/wahay"}
	osArgs = mockOsArgs

	md := &mockDir{}
	defer gostub.New().Stub(&filepathDir, md.Dir).Reset()
	md.On("Dir", osArgs[0]).Return(localDirectory).Once()

	ma := &mockAbs{}
	defer gostub.New().Stub(&filepathAbs, ma.Abs).Reset()
	ma.On("Abs", localDirectory).Return("", errors.New("Abs error")).Once()

	binary, err := searchBinaryInLocalDir()
	c.Assert(binary, IsNil)
	c.Assert(err, IsNil)

	md.AssertExpectations(c)
	ma.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInLocalDir_returnsAnInvalidBinaryWhenTheProgramNameIsEmptyAndTheCurrentDirectoryDoesNotHaveAMumbleBinary(c *C) {
	localDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(localDirectory)

	mockOsArgs := []string{""}
	osArgs = mockOsArgs

	md := &mockDir{}
	defer gostub.New().Stub(&filepathDir, md.Dir).Reset()
	md.On("Dir", osArgs[0]).Return(localDirectory).Once()

	ma := &mockAbs{}
	defer gostub.New().Stub(&filepathAbs, ma.Abs).Reset()
	ma.On("Abs", localDirectory).Return(localDirectory, nil).Once()

	binary, _ := searchBinaryInLocalDir()
	c.Assert(binary.isValid, IsFalse)
	c.Assert(binary.lastError, ErrorMatches, "not valid binary path")

	md.AssertExpectations(c)
	ma.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInLocalDir_returnsAnInvalidBinaryWhenThereIsNotAMumbleBinaryInTheLocalDirectory(c *C) {
	localDirectory, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(localDirectory)

	mockOsArgs := []string{localDirectory + "/wahay"}
	osArgs = mockOsArgs

	md := &mockDir{}
	defer gostub.New().Stub(&filepathDir, md.Dir).Reset()
	md.On("Dir", osArgs[0]).Return(localDirectory).Once()

	ma := &mockAbs{}
	defer gostub.New().Stub(&filepathAbs, ma.Abs).Reset()
	ma.On("Abs", localDirectory).Return(localDirectory, nil).Once()

	binary, _ := searchBinaryInLocalDir()
	c.Assert(binary.isValid, IsFalse)
	c.Assert(binary.lastError, ErrorMatches, "not valid binary path")

	md.AssertExpectations(c)
	ma.AssertExpectations(c)
}

type mockXdgDataHome struct {
	mock.Mock
}

func (m *mockXdgDataHome) XdgDataHome() string {
	args := m.Called()
	return args.String(0)
}

func (s *clientSuite) Test_searchBinaryInDataDir_returnsAValidBinaryIfABinaryFileIsFoundInTheDataDirectory(c *C) {
	dataDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dataDir)

	mumbleBinaryPath := filepath.Join(dataDir, "/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	binaryFile, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	mx := &mockXdgDataHome{}
	defer gostub.New().Stub(&configXdgDataHome, mx.XdgDataHome).Reset()
	mx.On("XdgDataHome").Return(dataDir)

	mj := &mockJoin{}
	defer gostub.New().Stub(&filepathJoin, mj.Join).Reset()
	mj.On("Join", []string{dataDir, mumbleBundlePath}).Return(binaryFile.Name())
	mj.On("Join", []string{dataDir, wahayMumbleBundlePath}).Return(filepath.Join(dataDir, wahayMumbleBundlePath))

	mc := &mockCommand{}
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()
	mc.On("Command", binaryFile.Name(), []string{"--version"}).Return(&exec.Cmd{Path: binaryFile.Name(), Args: []string{"--version"}}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary, err := searchBinaryInDataDir()
	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.lastError, IsNil)
	c.Assert(err, IsNil)

	mx.AssertExpectations(c)
	mj.AssertExpectations(c)
	mc.AssertExpectations(c)
}

func (s *clientSuite) Test_searchBinaryInDataDir_returnsNilWhenABinaryDoesNotExistInTheDataDirectory(c *C) {
	dataDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dataDir)

	mx := &mockXdgDataHome{}
	defer gostub.New().Stub(&configXdgDataHome, mx.XdgDataHome).Reset()
	mx.On("XdgDataHome").Return(dataDir)

	binary, err := searchBinaryInDataDir()
	c.Assert(binary, IsNil)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_envIfBundle_returnsNilIfTheBinaryIsNotBundled(c *C) {
	binary := &binary{isBundle: false}

	envVariables := binary.envIfBundle()
	c.Assert(envVariables, IsNil)
}

func (s *clientSuite) Test_envIfBundle_returnsTheRequiredMumbleBinaryEnvironmentVariables(c *C) {
	binary := &binary{isBundle: true, env: []string{"enviroment variable"}}

	envVariables := binary.envIfBundle()
	c.Assert(envVariables, DeepEquals, []string{"enviroment variable"})
}

func (s *clientSuite) Test_envIfBundle_returnsNilWhenBinaryIsBundledButDoesNotHaveEnvironmentVariables(c *C) {
	binary := &binary{isBundle: true, env: []string{}}

	envVariables := binary.envIfBundle()
	c.Assert(envVariables, IsNil)
	c.Assert(binary.isBundle, IsFalse)
}

func (s *clientSuite) Test_searchBinary_returnsNilWhenNoBinaryIsFound(c *C) {
	conf := &config.ApplicationConfig{}

	ml := &mockLookPath{}
	defer gostub.New().Stub(&execLookPath, ml.LookPath).Reset()
	ml.On("LookPath", "mumble").Return("", nil).Once()

	binary := searchBinary(conf)
	c.Assert(binary, IsNil)

	ml.AssertExpectations(c)
}
