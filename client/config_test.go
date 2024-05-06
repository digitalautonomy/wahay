package client

import (
	. "gopkg.in/check.v1"
)

func (s *clientSuite) Test_pathToConfig_returnsClientConfigDirIfIsAlreadySetted(c *C) {
	client := &client{configDir: "path/to/config"}

	pathToConfig := client.pathToConfig()

	c.Assert(pathToConfig, Equals, "path/to/config")

}
