package client

import (
	"os"
	"path/filepath"

	. "github.com/digitalautonomy/wahay/test"
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_readerMumbleDB_returnsTheByteRepresentationFromAString(c *C) {
	result := string(readerMumbleDB())

	c.Assert(result, HasLen, 122880)
	c.Assert(result, Contains, "SQLite format 3")
	c.Assert(result, Contains, "ffaaffaabbddaabbddeeaaddccaaffeebbaabbeeddeeaaddbbeeeeff.onion")
	c.Assert(result, Contains, "#indexpingcache_host_portpingcach")
}

func (s *clientSuite) Test_readerMumbleIniConfig_returnsTheContentLikeAString(c *C) {
	result := readerMumbleIniConfig()

	c.Assert(result, HasLen, 467)
	c.Assert(result, Contains, "version=1.3.0")
	c.Assert(result, Contains, "#CERTIFICATE")
	c.Assert(result, Contains, "#LANGUAGE")
}

func (s *clientSuite) Test_pathExists_returnsTrueWhenPathExists(c *C) {
	existingDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}

	defer os.RemoveAll(existingDir)

	result := pathExists(existingDir)

	c.Assert(result, IsTrue)
}

func (s *clientSuite) Test_pathExists_returnsFalseWhenPathDoesNotExists(c *C) {
	nonExistingDir := "/path/that/does/not/exist"

	result := pathExists(nonExistingDir)

	c.Assert(result, IsFalse)
}

func (s *clientSuite) Test_pathExists_returnsFalseWhenPathIsEmpty(c *C) {
	emptyPath := ""

	result := pathExists(emptyPath)

	c.Assert(result, IsFalse)
}

func (s *clientSuite) Test_isADirectory_returnsTrueWhenIsADirectoryAndFalseOtherwise(c *C) {
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

	result := isADirectory(existingDir)

	c.Assert(result, IsTrue)

	nonExistingDir := filepath.Join(tempDir, "non_existing_dir")
	result = isADirectory(nonExistingDir)

	c.Assert(result, IsFalse)

}

func (s *clientSuite) Test_isADirectory_returnsTrueWhenIsAFileAndFalseOtherwise(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	existingFile := filepath.Join(tempDir, "existing.txt")
	err = os.WriteFile(existingFile, []byte("file content"), 0644)
	if err != nil {
		c.Fatalf("Failed to create test file: %v", err)
	}

	result := isAFile(existingFile)

	c.Assert(result, IsTrue)

	nonExistingFile := filepath.Join(tempDir, "non_existing_file.txt")
	result = isAFile(nonExistingFile)

	c.Assert(result, IsFalse)

}

func (s *clientSuite) Test_createFile_actuallyCreatesAfile(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	newFilePath := filepath.Join(tempDir, "new_file.txt")
	err = createFile(newFilePath)
	if err != nil {
		c.Errorf("Expected no error when creating a new file, but got: %v", err)
	}

	//Check if the file exists
	_, err = os.Stat(newFilePath)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_createFile_worksWhenTheFileAlreadyExists(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	newFilePath := filepath.Join(tempDir, "new_file.txt")
	err = createFile(newFilePath)
	if err != nil {
		c.Errorf("Expected no error when creating a new file, but got: %v", err)
	}

	err = createFile(newFilePath)
	if err != nil {
		c.Errorf("Expected no error when creating a new file, but got: %v", err)
	}

	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_createFile_returnsAnErroWhenTheDirectoryPathIsInvalid(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidPath := filepath.Join(tempDir, "invalid/path/file.txt")
	err = createFile(invalidPath)

	c.Assert(err, NotNil)
}

func (s *clientSuite) Test_createDir_createsANewDirectoryWhenACorrectPathIsGiven(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	newDirPath := filepath.Join(tempDir, "new_dir")
	err = createDir(newDirPath)
	if err != nil {
		c.Errorf("Expected no error when creating a new directory, but got: %v", err)
	}

	//Check if the directory exists
	_, err = os.Stat(newDirPath)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_createDir_worksWhenTheDirectoryAlreadyExists(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	newDirPath := filepath.Join(tempDir, "new_dir")
	err = createDir(newDirPath)
	if err != nil {
		c.Errorf("Expected no error when creating a new directory, but got: %v", err)
	}

	err = createDir(newDirPath)
	if err != nil {
		c.Errorf("Expected no error when creating a new directory, but got: %v", err)
	}

	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_createDir_createsANestedDirectory(c *C) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		c.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	nestedDirPath := filepath.Join(tempDir, "nested/dir")
	err = createDir(nestedDirPath)
	if err != nil {
		c.Errorf("Expected no error when creating a new directory, but got: %v", err)
	}

	//Check if the directory exists
	_, err = os.Stat(nestedDirPath)
	c.Assert(err, IsNil)
}
