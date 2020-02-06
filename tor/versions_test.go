package tor

import (
	. "gopkg.in/check.v1"
)

type WahayTorVersionsSuite struct{}

var _ = Suite(&WahayTorVersionsSuite{})

func (s *WahayTorVersionsSuite) Test_parseVersion_returnsTheThreeImportantVersionNumbers(c *C) {
	maj, min, patch, err := parseVersion("0.12.3.4")
	c.Assert(err, IsNil)
	c.Assert(maj, Equals, 0)
	c.Assert(min, Equals, 12)
	c.Assert(patch, Equals, 3)
}

func (s *WahayTorVersionsSuite) Test_parseVersion_returnsErrorForBadNumberInTheFirstThreeParts(c *C) {
	_, _, _, err := parseVersion("x.12.3.4")
	c.Assert(err, ErrorMatches, "invalid version number")

	_, _, _, err = parseVersion("0.y.3.4")
	c.Assert(err, ErrorMatches, "invalid version number")

	_, _, _, err = parseVersion("0.12.z.4")
	c.Assert(err, ErrorMatches, "invalid version number")

	_, _, _, err = parseVersion("0.12.4.x")
	c.Assert(err, IsNil)
}

func (s *WahayTorVersionsSuite) Test_parseVersion_returnsErrorIfThereArentNumbers(c *C) {
	_, _, _, err := parseVersion("1.12")
	c.Assert(err, ErrorMatches, "invalid version string")

	_, _, _, err = parseVersion("1abc")
	c.Assert(err, ErrorMatches, "invalid version string")
}

func (s *WahayTorVersionsSuite) Test_compareVersions_comparesVersionsCorrectly(c *C) {
	cmp, _ := compareVersions("12.0.1.x", "9.1.42")
	c.Assert(cmp, Equals, 1)

	cmp, _ = compareVersions("8.2.1.x", "9.1.42")
	c.Assert(cmp, Equals, -1)

	cmp, _ = compareVersions("9.1.43.x", "9.3.42")
	c.Assert(cmp, Equals, -1)

	cmp, _ = compareVersions("9.1.41.x", "9.1.42")
	c.Assert(cmp, Equals, -1)

	cmp, _ = compareVersions("9.1.42.x", "9.1.42.z")
	c.Assert(cmp, Equals, 0)
}

func (s *WahayTorVersionsSuite) Test_compareVersions_returnsErrorForInvalidVersionString(c *C) {
	_, e := compareVersions("12.x.1.x", "9.1.42")
	c.Assert(e, ErrorMatches, "invalid version string")

	_, e = compareVersions("12.1.1.1", "x.1.42")
	c.Assert(e, ErrorMatches, "invalid version string")
}
