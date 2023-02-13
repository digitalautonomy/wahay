package tor

import (
	. "github.com/digitalautonomy/wahay/test"
	. "gopkg.in/check.v1"
)

type torSuite struct{}

var _ = Suite(&torSuite{})

func (s *torSuite) Test_getTorrc_returnsTheContentLikeAString(c *C) {
	content := getTorrc()

	c.Assert(content, HasLen, 558)
	c.Assert(content, Contains, "SOCKSPort __PORT__")
	c.Assert(content, Contains, "DataDirectory __DATADIR__")
	c.Assert(content, Contains, "CookieAuthentication __COOKIE__")
}

func (s *torSuite) Test_getTorrcLogConfig_returnsTheLogsContentLikeAString(c *C) {
	content := getTorrcLogConfig()

	c.Assert(content, HasLen, 177)
	c.Assert(content, Contains, "Log notice file __LOGNOTICE__")
	c.Assert(content, Contains, "Log debug file __LOGDEBUG__")
}
