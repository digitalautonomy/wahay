package client

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/digitalautonomy/wahay/test"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func createNestedDirsForTesting(baseDir string, dirNames []string, perm os.FileMode) (string, error) {
	currentDir := baseDir
	for _, dir := range dirNames {
		currentDir = filepath.Join(currentDir, dir)
		err := os.Mkdir(currentDir, perm)
		if err != nil {
			return "", fmt.Errorf("failed to create directory %s: %v", currentDir, err)
		}
	}
	return currentDir, nil
}

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

	nestedDirs := []string{"mumble", "mumble"}
	_, err = createNestedDirsForTesting(tempDir, nestedDirs, 0755)
	if err != nil {
		c.Fatalf("Failed to create nested directories: %v", err)
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

	nestedDirs := []string{"mumble", "mumble"}
	mumblePath, err := createNestedDirsForTesting(tempDir, nestedDirs, 0755)
	if err != nil {
		c.Fatalf("Failed to create nested directories: %v", err)
	}

	mumbleBundlePath := filepath.Join(tempDir, "/mumble", mumbleBundleLibsDir)

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

	nestedDirs := []string{"mumble", "mumble"}
	mumbleBundlePath, err := createNestedDirsForTesting(tempDir, nestedDirs, 0755)
	if err != nil {
		c.Fatalf("Failed to create nested directories: %v", err)
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

	nestedDirs := []string{"mumble", "lib"}
	_, err = createNestedDirsForTesting(tempDir, nestedDirs, 0755)
	if err != nil {
		c.Fatalf("Failed to create nested directories: %v", err)
	}

	mumbleBinaryPath := filepath.Join(tempDir, "/mumble/mumble")
	err = os.MkdirAll(mumbleBinaryPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	mc := &mockCommand{}

	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	defer gostub.New().Stub(&commandOutput, mc.Output).Reset()

	mc.On("Command", mumbleBinaryPath, []string{"-h"}).Return(&exec.Cmd{Path: mumbleBinaryPath, Args: []string{"-h"}}).Once()

	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary := isThereAnAvailableBinary(tempDir)

	c.Assert(binary.isBundle, IsTrue)
	c.Assert(binary.isValid, IsTrue)
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

	mc.On("Command", mumbleBinaryPath, []string{"-h"}).Return(&exec.Cmd{Path: mumbleBinaryPath, Args: []string{"-h"}}).Once()

	mc.On("Output").Return([]byte("command output"), nil).Once()

	binary := isThereAnAvailableBinary(tempDir)

	c.Assert(binary.isValid, IsTrue)
	c.Assert(binary.isBundle, IsFalse)
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
}

type mockLookPath struct {
	mock.Mock
}

func (m *mockLookPath) LookPath(file string) (string, error) {
	args := m.Called(file)
	return args.String(0), args.Error(1)
}

func (s *clientSuite) Test_searchBinaryInSystem_returnsTheBinaryFoundInTheSystem(c *C) {
	ml := &mockLookPath{}
	defer gostub.New().Stub(&execLookPath, ml.LookPath).Reset()

	ml.On("LookPath", "mumble").Return("absolut/path/to/binary", nil).Once()

	binary, err := searchBinaryInSystem()

	c.Assert(binary, NotNil)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_searchBinaryInSystem_returnsNilWhenTheBinaryIsNotFoundInTheSystem(c *C) {
	ml := &mockLookPath{}
	defer gostub.New().Stub(&execLookPath, ml.LookPath).Reset()

	ml.On("LookPath", "mumble").Return("", errors.New("Error looking for executable")).Once()

	binary, _ := searchBinaryInSystem()
	c.Assert(binary, IsNil)
}
