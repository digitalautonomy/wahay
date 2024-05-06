package client

import (
	"os"
	"path/filepath"

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
