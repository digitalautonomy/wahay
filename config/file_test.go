package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (cs *ConfigSuite) Test_CreateTempDir_createsTempFileInWahayDirectory(c *C) {
	dir := CreateTempDir("test")

	tempDir := filepath.Dir(dir)

	c.Assert(dir, Not(Equals), "")
	c.Assert(tempDir, Equals, wahayDataDir)

	_, err := os.Stat(dir)
	c.Assert(err, IsNil)

	defer os.RemoveAll(dir)
}

func (cs *ConfigSuite) Test_FileExists_returnsTrueWhenFileExists(c *C) {
	tempFile, err := ioutil.TempFile("", "example")

	c.Assert(err, IsNil)
	defer os.Remove(tempFile.Name())

	exists := FileExists(tempFile.Name())
	c.Assert(exists, Equals, true)
}

func (cs *ConfigSuite) Test_FileExists_returnsFalseWhenFileDoesNotExist(c *C) {
	tempFile := filepath.Join(os.TempDir(), "non_existent_file.txt")

	exists := FileExists(tempFile)
	c.Assert(exists, Equals, false)

	defer os.Remove(tempFile)
}

func (cs *ConfigSuite) Test_EnsureDir_createsDirWhenDirectoryDoesNotExist(c *C) {
	tempDir, err := ioutil.TempDir("", "testdir")

	c.Assert(err, IsNil)
	defer os.Remove(tempDir)

	dirToCreate := filepath.Join(tempDir, "new_directory")

	EnsureDir(dirToCreate, 0700)

	_, err = os.Stat(dirToCreate)
	c.Assert(err, IsNil)
}

func (cs *ConfigSuite) Test_EnsureDir_doesNothingIfDirectoryAlreadyExists(c *C) {
	tempDir, err := ioutil.TempDir("", "testdir")

	c.Assert(err, IsNil)
	defer os.Remove(tempDir)

	dirToCreate := filepath.Join(tempDir, "new_directory")

	EnsureDir(dirToCreate, 0700)

	_, err = os.Stat(dirToCreate)
	c.Assert(err, IsNil)
}

func (cs *ConfigSuite) Test_SafeWrite_writesContentOnTempFile(c *C) {
	tempDir, err := ioutil.TempDir("", "testdir")

	c.Assert(err, IsNil)
	defer os.Remove(tempDir)

	fileName := filepath.Join(tempDir, "test.txt")
	fileContent := []byte("Praise the sun!")

	err = SafeWrite(fileName, fileContent, 0644)
	c.Assert(err, IsNil)

	content, err := ioutil.ReadFile(fileName)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, fileContent)

	_, err = os.Stat(fileName)
	c.Assert(err, IsNil)
}
