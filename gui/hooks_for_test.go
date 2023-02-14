package gui

import (
	log "github.com/sirupsen/logrus"
	"io"
)

func init() {
	log.SetOutput(io.Discard)
}
