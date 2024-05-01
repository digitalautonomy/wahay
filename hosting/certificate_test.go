package hosting

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/mock"
	. "gopkg.in/check.v1"
)

func (s *hostingSuite) Test_newCertificateServer_generatesNewCertificatedWebServerSuccesfully(c *C) {
	path := "/tmp/wahay"
	file := "cert.pem"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(path, perm)

	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)

	fp := filepath.Join(path, file)
	_, e = os.Create(fp)

	if e != nil {
		c.Fatalf("Failed to create file: %v", e)
	}

	httpServer, err := newCertificateServer(path)

	c.Assert(httpServer, NotNil)
	c.Assert(httpServer.address, Matches, `127.0.0.1:.*`)
	c.Assert(httpServer.cert, DeepEquals, []byte{})
	c.Assert(httpServer.server.Addr, Equals, httpServer.address)
	c.Assert(httpServer.server.ReadTimeout, Equals, 5*time.Second)
	c.Assert(httpServer.server.WriteTimeout, Equals, 10*time.Second)
	c.Assert(httpServer.server.IdleTimeout, Equals, 120*time.Minute)
	c.Assert(err, IsNil)
}

func (s *hostingSuite) Test_newCertificateServer_returnsAnErrorWhenNoCertificateFileExistsOnPath(c *C) {
	path := "/tmp/wahay"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(path, perm)

	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	defer os.RemoveAll(path)
	httpServer, err := newCertificateServer(path)
	expectedErr := "the certificate file do not exists"

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, expectedErr)
	c.Assert(httpServer, IsNil)
}

func (s *hostingSuite) Test_newCertificate_returnsAnErrorWhenDirectoryDoNotExists(c *C) {
	path := "fake/dir"
	expectedErr := "the certificate file do not exists"

	httpServer, err := newCertificateServer(path)

	c.Assert(httpServer, IsNil)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, expectedErr)
}

type mockIoutil struct {
	mock.Mock
}

func (m *mockIoutil) ReadFile(file string) ([]byte, error) {
	ret := m.Called(file)
	return ret.Get(0).([]byte), ret.Error(1)
}

func (s *hostingSuite) Test_newCertificate_returnsAnErrorWhenCertificateContentCantBeReaded(c *C) {
	path := "/tmp/wahay"
	var perm fs.FileMode = 0700
	e := os.MkdirAll(path, perm)

	if e != nil {
		c.Fatalf("Failed to create temporary directory: %v", e)
	}

	file := "cert.pem"
	fp := filepath.Join(path, file)
	_, e = os.Create(fp)

	if e != nil {
		c.Fatalf("Failed to create file: %v", e)
	}

	defer os.RemoveAll(path)

	mi := &mockIoutil{}
	defer gostub.New().Stub(&ioutilReadFile, mi.ReadFile).Reset()
	var ea []byte
	expectedErr := errors.New("open " + fp + ": no such file or directory")
	mi.On("ReadFile", fp).Return(ea, expectedErr)

	httpServer, err := newCertificateServer(path)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, expectedErr.Error())
	c.Assert(httpServer, IsNil)
}

func (s *hostingSuite) Test_stop_returnsNoErrorWhenWebServerIsNotRunning(c *C) {
	ws := &webserver{
		running: false,
	}
	err := ws.stop()

	c.Assert(err, IsNil)
	c.Assert(ws.running, Equals, false)
}

func (s *hostingSuite) Test_stop_worksWithBasicExample(c *C) {
	ws := &webserver{
		running: true,
		server:  &http.Server{},
	}
	err := ws.stop()

	c.Assert(err, IsNil)
	c.Assert(ws.running, Equals, false)
}

func (s *hostingSuite) Test_start_keepsServerRunningWhenItHasAlreadyStarted(c *C) {
	ws := &webserver{
		running: true,
	}
	ws.start(func(err error) {})
	c.Assert(ws.running, Equals, true)
}
