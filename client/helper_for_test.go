package client

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type clientSuite struct{}

var _ = Suite(&clientSuite{})

func init() {
	logrus.SetOutput(io.Discard)
}
