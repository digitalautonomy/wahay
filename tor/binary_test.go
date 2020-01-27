package tor

import (
	"log"

	. "gopkg.in/check.v1"
)

type TonioTorBinarySuite struct{}

var _ = Suite(&TonioTorBinarySuite{})

func (s *TonioTorBinarySuite) Test_ValidConfiguredPath(c *C) {
	log.Printf("Test_ValidConfiguredPath")

	var torBinaryPath = "/usr/bin/tor"
	pathBinTor := Initialize(torBinaryPath)
	c.Assert(pathBinTor, Equals, "/usr/bin/tor")
}

func (s *TonioTorBinarySuite) Test_WhichConfiguredPath(c *C) {
	log.Printf("Test_WhichConfiguredPath")

	var torBinaryPath = "/tmp/bin/tor"
	pathBinTor := Initialize(torBinaryPath)
	c.Assert(pathBinTor, Equals, "/usr/sbin/tor")
}
