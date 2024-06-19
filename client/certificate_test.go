package client

import (
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_extractHostAndPort_returnsAnErrorWhenFailsParsingAddress(c *C) {
	h, p, err := extractHostAndPort("://")
	c.Assert(h, Equals, "")
	c.Assert(p, Equals, "")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "parse \"://\": missing protocol scheme")
}

func (s *clientSuite) Test_extractHostAndPort_returnsAnErrorWhenFailsSplittingAddress(c *C) {
	h, p, err := extractHostAndPort("test")
	c.Assert(h, Equals, "")
	c.Assert(p, Equals, "")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "missing port in address")
}

func (s *clientSuite) Test_extractHostAndPort_returnsAnErrorWhenPortIsMissing(c *C) {
	add := "http://test"
	_, _, err := extractHostAndPort(add)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "address test: missing port in address")
}

func (s *clientSuite) Test_extractHostAndPort_returnsAnErrorWhenHostIsMi(c *C) {
	add := ":420"
	_, _, err := extractHostAndPort(add)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "parse \":420\": missing protocol scheme")
}

func (s *clientSuite) Test_extractHostAndPort_returnsHostAndPortWhenAddressIsCorrect(c *C) {
	h, p, err := extractHostAndPort("http://test:666")
	c.Assert(h, Equals, "test")
	c.Assert(p, Equals, "666")
	c.Assert(err, IsNil)
}

func (s *clientSuite) Test_requestCertificate_returnsAnErrorWhenFailsExtractingHostAndPort(c *C) {
	cl := client{}
	err := cl.requestCertificate("http://test")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "invalid certificate url")
}
