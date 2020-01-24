package tor

import (
	. "gopkg.in/check.v1"
)

type TonioTorBinarySuite struct{}

var _ = Suite(&TonioTorBinarySuite{})

func (s *TonioTorBinarySuite) Test_CheckBinaryDefaultPathsLoaded(c *C) {
	var torBinaryPaths = make([]string, 0)
	torBinary = nil
	b := GetTorBinary(torBinaryPaths)
	c.Assert(len(b.GetPaths()), Equals, 3)
}

func (s *TonioTorBinarySuite) Test_CheckBinaryOnePathLoaded(c *C) {
	var torBinaryPaths = []string{"/tmp"}
	torBinary = nil
	b := GetTorBinary(torBinaryPaths)
	c.Assert(len(b.GetPaths()), Equals, 1)
}

func (s *TonioTorBinarySuite) Test_CheckBinaryTwoPathsLoaded(c *C) {
	var torBinaryPaths = []string{"/tmp", "/opt"}
	torBinary = nil
	b := GetTorBinary(torBinaryPaths)
	c.Assert(len(b.GetPaths()), Equals, 2)
}

func (s *TonioTorBinarySuite) Test_InvalidTorPath(c *C) {
	var torBinaryPaths = []string{"/tmp"}
	torBinary = nil
	b := GetTorBinary(torBinaryPaths)
	c.Assert(b.Check(), ErrorMatches, "no Tor binary found")
}

/*
func (s *TonioTorBinarySuite) Test_ValidTorPath(c *C) {
	var torBinaryPaths = []string{"/usr/bin/tor"}
	torBinary = nil
	b := GetTorBinary(torBinaryPaths)
	err := b.Check()
	c.Assert(err, IsNil)
}*/
