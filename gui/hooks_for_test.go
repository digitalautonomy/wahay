package gui

import (
	"io"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetOutput(io.Discard)
}
