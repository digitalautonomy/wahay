package config

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

func init() {
	logrus.SetOutput(io.Discard)
}
