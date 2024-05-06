package client

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_pathToConfig_returnsClientConfigDirIfIsAlreadySetted(c *C) {
	client := &client{configDir: "path/to/config"}

	pathToConfig := client.pathToConfig()

	c.Assert(pathToConfig, Equals, "path/to/config")

}

func (s *clientSuite) Test_pathToConfig_returnsAndSetsParentDirOfBinaryIfConfigDirIsEmptyAndPathIsBinary(c *C) {
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

	binaryFile, err := os.CreateTemp(mumbleBinaryPath, "mumble")
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	client := &client{isValid: true, binary: &binary{path: binaryFile.Name()}, configDir: ""}

	pathToConfig := client.pathToConfig()
	expectedPath := filepath.Dir(binaryFile.Name())
	c.Assert(pathToConfig, Equals, expectedPath)
	c.Assert(client.configDir, Equals, expectedPath)
}

func (s *clientSuite) Test_pathToConfig_returnsAndSetsBinaryPathIfConfigDirIsEmptyAndPathIsDirectory(c *C) {
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

	client := &client{isValid: true, binary: &binary{path: mumbleBinaryPath}, configDir: ""}

	pathToConfig := client.pathToConfig()
	expectedPath := mumbleBinaryPath
	c.Assert(pathToConfig, Equals, expectedPath)
	c.Assert(client.configDir, Equals, expectedPath)
}

func (s *clientSuite) Test_ensureConfigurationDir_returnsNilAndCreatesConfigDirAndSubdirs(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configurationPath := filepath.Join(tempDir, "/mumble")
	err = os.MkdirAll(configurationPath, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	client := &client{isValid: true, binary: &binary{path: configurationPath}, configDir: ""}

	err = client.ensureConfigurationDir()
	c.Assert(err, IsNil)

	// Assert that the Configuration directory exists.
	_, err = os.Stat(configurationPath)
	c.Assert(err, IsNil)

	//Assert that required configuration subdirectories are created.
	for _, dir := range mumbleFolders {
		_, err = os.Stat(filepath.Join(configurationPath, dir))
		c.Assert(err, IsNil)
	}
}

type mockMkdirAll struct {
	mock.Mock
}

func (m *mockMkdirAll) MkdirAll(path string, perm fs.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (s *clientSuite) Test_ensureConfigurationDir_returnsAnErrorIfCreateDirFails(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := &client{isValid: true, configDir: "path/to/config"}

	mm := &mockMkdirAll{}
	defer gostub.New().Stub(&osMkdirAll, mm.MkdirAll).Reset()
	var perm fs.FileMode = 0700
	mm.On("MkdirAll", "path/to/config", perm).Return(errors.New("Error creating directory")).Once()

	err = client.ensureConfigurationDir()
	c.Assert(err, ErrorMatches, "Error creating directory")
}
