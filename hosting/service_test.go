package hosting

import (
	. "gopkg.in/check.v1"
)

func (s *hostingSuite) Test_defaultHost_returnsLocalHostInterfaceWhenWorkstationFileHasNotBeenFound(c *C) {
	dh := defaultHost()

	c.Assert(dh, Equals, "127.0.0.1")
}
