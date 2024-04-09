package client

import (
	"errors"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

type mockTempDir struct {
	mock.Mock
}

func (m *mockTempDir) tempDir(dir, prefix string) (string, error) {
	args := m.Called(dir, prefix)
	return args.String(0), args.Error(1)
}

func (s *clientSuite) Test_tempFolder_createsATempFolderSuccessfully(c *C) {
	tempDir, err := tempFolder()

	c.Assert(tempDir, NotNil)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_tempFolder_returnsAnErrorWhenTheTempFolderIsNotCreated(c *C) {

	mtd := &mockTempDir{}

	defer gostub.New().Stub(&tempDir, mtd.tempDir).Reset()

	mtd.On("tempDir", "", "mumble").Return("", errors.New("Error creating the temp folder")).Once()
	tempDir, err := tempFolder()

	c.Assert(tempDir, Equals, "")
	c.Assert(err, NotNil)
}

func (s *clientSuite) Test_LastError_returnsClientError(c *C) {
	client := &client{err: errors.New("Client Error")}

	result := client.LastError()

	c.Assert(result, NotNil)
}

func (s *clientSuite) Test_pathToBinary_returnsTheValidBinaryPath(c *C) {
	client := &client{isValid: true, binary: &binary{path: "path/to/binary"}}

	result := client.pathToBinary()

	c.Assert(result, Equals, "path/to/binary")
}
