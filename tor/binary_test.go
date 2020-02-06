package tor

import (
	. "gopkg.in/check.v1"
)

type WahayTorBinarySuite struct{}

var _ = Suite(&WahayTorBinarySuite{})

/*func (s *WahayTorBinarySuite) Test_ValidConfiguredPath(c *C) {
	log.Printf("Test_ValidConfiguredPath")

	var torBinaryPath = "/usr/bin/tor"
	pathBinTor := Initialize(torBinaryPath)
	c.Assert(pathBinTor, Equals, "/usr/bin/tor")
}*/

/*func (s *WahayTorBinarySuite) Test_WhichConfiguredPath(c *C) {
	log.Printf("Test_WhichConfiguredPath")

	var torBinaryPath = "/tmp/bin/tor"
	pathBinTor := Initialize(torBinaryPath)
	c.Assert(pathBinTor, Equals, "/usr/sbin/tor")
}*/
