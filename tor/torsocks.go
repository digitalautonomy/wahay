package tor

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrTorsocksNotInstalled is an error to be trown where
	// torsocks is not installed in the system
	ErrTorsocksNotInstalled = errors.New("torsocks not available")
)

func findTorsocksBinary() (fatalErr error) {
	return findTorsocksInSystem()
}

func findTorsocksInSystem() (fatalErr error) {
	path, err := execf.LookPath("torsocks")
	if err != nil {
		log.Errorf("Torsocks is not installed in your system: %s", err.Error())
		return ErrTorsocksNotInstalled
	}

	log.Debugf("findTorsocksInSystem(%s)", path)

	return nil
}
