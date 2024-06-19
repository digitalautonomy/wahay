package client

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

// helper functions
func createTempDir(c *C) string {
	tempDir, err := ioutil.TempDir("", "test")
	c.Assert(err, IsNil)
	return tempDir
}

func createTempFile(c *C, dir, filename, content string) string {
	tempFile := filepath.Join(dir, filename)
	err := ioutil.WriteFile(tempFile, []byte(content), 0600)
	c.Assert(err, IsNil)
	return tempFile
}

func removeTempDir(c *C, dir string) {
	err := os.RemoveAll(dir)
	c.Assert(err, IsNil)
}

// tests
func (s *clientSuite) Test_write_writesDBContent(c *C) {
	tempDir := createTempDir(c)
	tempFile := createTempFile(c, tempDir, "testfile.txt", "")
	defer removeTempDir(c, tempDir)

	file, err := os.Create(tempFile)
	c.Assert(err, IsNil)
	defer file.Close()

	content := []byte("example content")

	db := &dbData{
		filename: tempFile,
		content:  content,
	}

	err = db.write()
	c.Assert(err, IsNil)

	readContent, err := ioutil.ReadFile(tempFile)

	c.Assert(err, IsNil)
	c.Assert(readContent, DeepEquals, content)
}

func (s *clientSuite) Test_write_handlesError(c *C) {
	tempDir, err := ioutil.TempDir("", "test")
	c.Assert(err, IsNil)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "testfile.txt")

	file, err := os.Create(tempFile)
	c.Assert(err, IsNil)
	file.Close()

	err = os.Chmod(tempFile, 0400)
	c.Assert(err, IsNil)

	content := []byte("example content")

	db := &dbData{
		filename: tempFile,
		content:  content,
	}

	err = db.write()

	c.Assert(err, NotNil)

}

func (s *clientSuite) Test_exists_returnsTrueIfStringExistsInContent(c *C) {
	db := &dbData{
		content: []byte("this is true"),
	}

	exampleString := "true"

	res := db.exists(exampleString)

	c.Assert(res, Equals, true)
}

func (s *clientSuite) Test_exists_returnsFalseIfStringDoesntExistInContent(c *C) {
	db := &dbData{
		content: []byte("this is false"),
	}

	exampleString := "true"

	res := db.exists(exampleString)

	c.Assert(res, Equals, false)
}

func (s *clientSuite) Test_replaceString_findsAndReplacesContentInDB(c *C) {
	db := &dbData{
		content: []byte("this is true"),
	}

	find := "true"
	replace := "false"
	expectedContent := []byte("this is false")

	db.replaceString(find, replace)

	c.Assert(db.content, DeepEquals, expectedContent)
}

func (s *clientSuite) Test_readBinaryContent_readsFileContent(c *C) {
	tempDir, err := ioutil.TempDir("", "test")
	c.Assert(err, IsNil)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "testfile.bin")
	expectedContent := []byte("example binary content")

	defer os.Remove(tempFile)

	err = ioutil.WriteFile(tempFile, expectedContent, 0600)
	c.Assert(err, IsNil)

	actualContent, err := readBinaryContent(tempFile)
	c.Assert(err, IsNil)
	c.Assert(actualContent, DeepEquals, expectedContent)
}

func (s *clientSuite) Test_readBinaryContent_handlesFileNotFoundError(c *C) {
	nonExistentFile := "/path/to/nonexistent/file.bin"

	_, err := readBinaryContent(nonExistentFile)
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *clientSuite) Test_loadDBFromFile_loadsDatabaseSuccessfully(c *C) {
	tempDir, err := ioutil.TempDir("", "test")
	c.Assert(err, IsNil)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "testfile.db")
	expectedContent := []byte("example database content")

	defer os.Remove(tempFile)

	err = ioutil.WriteFile(tempFile, expectedContent, 0600)
	c.Assert(err, IsNil)

	db, err := loadDBFromFile(tempFile)
	c.Assert(err, IsNil)
	c.Assert(db, NotNil)
	c.Assert(db.filename, Equals, tempFile)
	c.Assert(db.content, DeepEquals, expectedContent)
}

func (s *clientSuite) Test_loadDBFromFile_handlesFileNotFoundError(c *C) {
	nonExistentFile := "/path/to/nonexistent/file.db"

	db, err := loadDBFromFile(nonExistentFile)
	c.Assert(err, NotNil)
	c.Assert(os.IsNotExist(err), Equals, true)
	c.Assert(db, IsNil)
}
