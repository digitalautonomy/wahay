package hosting

import (
	"io/fs"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (s *hostingSuite) Test_newCertificateServer_generatesAServerCertificateSuccesfully(c *C) {
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
