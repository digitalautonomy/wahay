package client

import (
	"errors"
	"os/exec"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_storeCertificate_returnsAnErrorWhenBadCertificateHasBeenGiven(c *C) {
	var mumbleDBContent []byte

	d := func() []byte {
		return mumbleDBContent
	}

	cl := client{
		databaseProvider: d,
	}
	cert := []byte("dummy cert")
	err := cl.storeCertificate("test", 123, cert)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "invalid certificate")
}

func (s *clientSuite) Test_storeCertificate_returnsNoErrorWhenSuccesfullyStoresCertificateInDB(c *C) {
	var mumbleDBContent []byte

	d := func() []byte {
		return mumbleDBContent
	}

	cl := client{
		databaseProvider: d,
	}

	cert := []byte(fakeCert)
	err := cl.storeCertificate("test", 123, cert)
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_generateTemporaryMumbleCertificate_returnsCertificateSuccessfully(c *C) {
	mc := &mockCommand{}
	defer gostub.New().Stub(&cmdOutput, mc.Output).Reset()
	mc.On("Command", "openssl", mock.Anything).Return(&exec.Cmd{}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()

	mrf := &mockReadFile{}
	defer gostub.New().Stub(&osReadFile, mrf.ReadFile).Reset()
	mrf.On("ReadFile", mock.Anything).Return([]byte("data content"), nil).Once()

	data, err := generateTemporaryMumbleCertificate()
	c.Assert(err, IsNil)
	c.Assert(data, NotNil)
	c.Assert(data, Matches, `@ByteArray\(data content\)`)

	mc.AssertExpectations(c)
	mrf.AssertExpectations(c)
}

func (s *clientSuite) Test_generateTemporaryMumbleCertificate_returnsAnErrorWhenFailsCreatingTempDir(c *C) {
	mtd := &mockTempDir{}
	defer gostub.New().Stub(&ioutilTempDir, mtd.tempDir).Reset()

	expectedError := "Error creating the temp folder"
	mtd.On("tempDir", "", "wahay_cert_generation").Return("", errors.New(expectedError)).Once()

	data, err := generateTemporaryMumbleCertificate()
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, expectedError)
	c.Assert(data, Equals, "")

	mtd.AssertExpectations(c)
}

func (s *clientSuite) Test_generateTemporaryMumbleCertificate_returnsAnErrorWhenFailsRunningTheOpensslCommand(c *C) {
	mc := &mockCmd{}
	defer gostub.New().Stub(&cmdOutput, mc.Output).Reset()

	expectedError := errors.New("OpenSSL is not installed on the system.")
	mc.On("Output").Return([]byte(""), expectedError)

	data, err := generateTemporaryMumbleCertificate()

	c.Assert(data, Equals, "")
	c.Assert(err, NotNil)
	c.Assert(err, Equals, expectedError)

	mc.AssertExpectations(c)
}

type mockReadFile struct {
	mock.Mock
}

func (m *mockReadFile) ReadFile(name string) ([]byte, error) {
	args := m.Called(name)
	return args.Get(0).([]byte), args.Error(1)
}

func (s *clientSuite) Test_generateTemporaryMumbleCertificate_returnsAnErrorWhenCantReadCertFile(c *C) {
	mc := &mockCommand{}
	defer gostub.New().Stub(&cmdOutput, mc.Output).Reset()
	mc.On("Command", "openssl", mock.Anything).Return(&exec.Cmd{}).Once()
	defer gostub.New().Stub(&execCommand, mc.Command).Reset()
	mc.On("Output").Return([]byte("command output"), nil).Once()

	mrf := &mockReadFile{}
	defer gostub.New().Stub(&osReadFile, mrf.ReadFile).Reset()

	mrf.On("ReadFile", mock.Anything).Return([]byte{}, errors.New("error reading certificate file")).Once()

	data, err := generateTemporaryMumbleCertificate()

	c.Assert(data, Equals, "")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "error reading certificate file")

	mc.AssertExpectations(c)
	mrf.AssertExpectations(c)
}
