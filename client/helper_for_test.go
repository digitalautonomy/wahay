package client

import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type clientSuite struct{}

var _ = Suite(&clientSuite{})
