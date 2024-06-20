package config

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct{}

var _ = Suite(&ConfigSuite{})

func init() {
	logrus.SetOutput(io.Discard)
}
