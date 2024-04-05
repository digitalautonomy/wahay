package client

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/digitalautonomy/wahay/test"
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
	mumbleBundlePath, err := createNestedDirsForTesting(tempDir, nestedDirs, 0755)
	if err != nil {
		c.Fatalf("Failed to create nested directories: %v", err)
	}

	mumbleBundlePathWithDependencies := filepath.Join(tempDir, "/mumble", mumbleBundleLibsDir)

	err = os.MkdirAll(mumbleBundlePathWithDependencies, 0755)
	if err != nil {
		c.Fatalf("Failed to create directory: %v", err)
	}

	isBundle, _ := checkLibsDependenciesInPath(mumbleBundlePath)

	c.Assert(isBundle, IsTrue)
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

	isBundle, _ := checkLibsDependenciesInPath(mumbleBundlePath)

	c.Assert(isBundle, IsFalse)
}
