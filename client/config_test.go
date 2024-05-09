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

func (s *clientSuite) Test_writeConfigToFile_successfullyWritesConfigurationContentToAnExistingFileAndUpdatesConfigFile(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile, err := os.Create(filepath.Join(tempDir, configFileName))
	if err != nil {
		c.Fatalf("Failed to create file")
	}

	client := &client{configFile: configFile.Name(), configContentProvider: func() string { return "config file content" }}

	err = client.writeConfigToFile(configFile.Name())
	c.Assert(err, IsNil)
	c.Assert(client.configFile, Equals, configFile.Name())

	fileContent, err := os.ReadFile(configFile.Name())
	if err != nil {
		c.Fatalf("Failed to read file")
	}
	c.Assert(string(fileContent), Equals, "config file content")
}

func (s *clientSuite) Test_writeConfigToFile_successfullyWritesConfigurationContentToANewFileInAnExistingDirectoryAndUpdatesConfigFile(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := &client{configContentProvider: func() string { return "config file content" }}
	c.Assert(client.configFile, Equals, "")

	err = client.writeConfigToFile(tempDir)
	c.Assert(err, IsNil)

	expectedConfigFile := filepath.Join(tempDir, configFileName)
	c.Assert(client.configFile, Equals, expectedConfigFile)

	fileContent, err := os.ReadFile(expectedConfigFile)
	if err != nil {
		c.Fatalf("Failed to read file")
	}
	c.Assert(string(fileContent), Equals, "config file content")
}

type mockCreate struct {
	mock.Mock
}

func (m *mockCreate) Create(name string) (*os.File, error) {
	args := m.Called(name)
	return args.Get(0).(*os.File), args.Error(1)
}

func (s *clientSuite) Test_writeConfigToFile_returnsAnErrorWhenTheConfigFileCreationFails(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	client := &client{}

	mc := &mockCreate{}
	defer gostub.New().Stub(&osCreate, mc.Create).Reset()
	configFile := filepath.Join(tempDir, configFileName)
	mc.On("Create", configFile).Return(&os.File{}, errors.New("Error creating file")).Once()

	err = client.writeConfigToFile(tempDir)
	c.Assert(err, Equals, errInvalidConfigFileDBFile)
}

func (s *clientSuite) Test_ensureConfigurationDBFile_successfullyCreatesAndWritesTheConfigurationDBFile(c *C) {
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

	client := &client{configDir: configurationPath, databaseProvider: func() []byte { return []byte("database configuration content") }}

	err = client.ensureConfigurationDBFile()
	c.Assert(err, IsNil)

	expectedConfigDatabaseFile := filepath.Join(configurationPath, configDBName)
	fileContent, err := os.ReadFile(expectedConfigDatabaseFile)
	if err != nil {
		c.Fatalf("Failed to read file")
	}
	c.Assert(string(fileContent), Equals, "database configuration content")
}

func (s *clientSuite) Test_ensureConfigurationDBFile_returnsAnErrorWhenTheConfigurationDBFileCreationFails(c *C) {
	client := &client{configDir: "invalid/configuration/path"}

	err := client.ensureConfigurationDBFile()
	c.Assert(err, NotNil)

	configurationDBFile := filepath.Join(client.configDir, configDBName)

	//Assert that the configuration file does not exist.
	_, err = os.Stat(configurationDBFile)
	c.Assert(err, NotNil)
}

func (s *clientSuite) Test_ensureConfigurationFile_successfullyCreatesAndWritesAConfigurationFile(c *C) {
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

	client := &client{configDir: configurationPath, configContentProvider: func() string { return "config file content" }}

	err = client.ensureConfigurationFile()
	c.Assert(err, IsNil)

	expectedConfigFile := filepath.Join(configurationPath, configFileName)
	fileContent, err := os.ReadFile(expectedConfigFile)
	if err != nil {
		c.Fatalf("Failed to read file")
	}
	c.Assert(string(fileContent), Equals, "config file content")
}

func (s *clientSuite) Test_ensureConfigurationFile_returnsAnErrorWhenTheConfigurationFileCreationFails(c *C) {
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

	client := &client{configDir: configurationPath, configContentProvider: func() string { return "config file content" }}

	mc := &mockCreate{}
	defer gostub.New().Stub(&osCreate, mc.Create).Reset()
	configFile := filepath.Join(client.configDir, configFileName)
	mc.On("Create", configFile).Return(&os.File{}, errors.New("Error creating file")).Once()

	err = client.ensureConfigurationFile()
	c.Assert(err, Equals, errInvalidConfigFileDBFile)
}
