package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type FileSuite struct{}

var _ = Suite(&FileSuite{})

func (f *FileSuite) Test_CreateTempDir_createsTempFileInWahayDirectory(c *C) {
	dir := CreateTempDir("test")

	tempDir := filepath.Dir(dir)

	c.Assert(dir, Not(Equals), "")
	c.Assert(tempDir, Equals, wahayDataDir)

	_, err := os.Stat(dir)
	c.Assert(err, IsNil)

	defer os.RemoveAll(dir)
}

func (f *FileSuite) Test_FileExists_returnsTrueWhenFileExists(c *C) {
	tempFile, err := ioutil.TempFile("", "example")

	c.Assert(err, IsNil)
	defer os.Remove(tempFile.Name())

	exists := FileExists(tempFile.Name())
	c.Assert(exists, Equals, true)
}

func (f *FileSuite) Test_FileExists_returnsFalseWhenFileDoesNotExist(c *C) {
	tempFile := filepath.Join(os.TempDir(), "non_existent_file.txt")

	exists := FileExists(tempFile)
	c.Assert(exists, Equals, false)
}

func (f *FileSuite) Test_EnsureDir_createsDirWhenDirectoryDoesNotExist(c *C) {
	tempDir := c.MkDir()

	dirToCreate := filepath.Join(tempDir, "new_directory")

	EnsureDir(dirToCreate, 0700)

	_, err := os.Stat(dirToCreate)
	c.Assert(err, IsNil)
}

func (f *FileSuite) Test_EnsureDir_doesNothingIfDirectoryAlreadyExists(c *C) {
	tempDir := c.MkDir()

	dirToCreate := filepath.Join(tempDir, "existing_directory")
	err := os.Mkdir(dirToCreate, 0700)
	c.Assert(err, IsNil)

	EnsureDir(dirToCreate, 0700)

	_, err = os.Stat(dirToCreate)
	c.Assert(err, IsNil)
}

func (f *FileSuite) Test_SafeWrite_writesContentOnTempFile(c *C) {
	testDir := c.MkDir()

	fileName := filepath.Join(testDir, "test.txt")
	fileContent := []byte("Praise the sun!")

	err := SafeWrite(fileName, fileContent, 0644)
	c.Assert(err, IsNil)

	content, err := ioutil.ReadFile(fileName)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, fileContent)

	_, err = os.Stat(fileName)
	c.Assert(err, IsNil)
}