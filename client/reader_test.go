package client

import . "gopkg.in/check.v1"

func (s *clientSuite) Test_readerMumbleDB_returnsTheByteRepresentationFromAString(c *C) {
	result := string(readerMumbleDB())

	c.Assert(result, HasLen, 122880)
	c.Assert(result, Contains, "SQLite format 3")
	c.Assert(result, Contains, "ffaaffaabbddaabbddeeaaddccaaffeebbaabbeeddeeaaddbbeeeeff.onion")
	c.Assert(result, Contains, "#indexpingcache_host_portpingcach")
}

func (s *clientSuite) Test_readerMumbleIniConfig_returnsTheContentLikeAString(c *C) {
	result := readerMumbleIniConfig()

	c.Assert(result, HasLen, 467)
	c.Assert(result, Contains, "version=1.3.0")
	c.Assert(result, Contains, "#CERTIFICATE")
	c.Assert(result, Contains, "#LANGUAGE")
}
