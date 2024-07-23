//go:build !windows

package tor

import (
	log "github.com/sirupsen/logrus"
)

func searchProxyTool() error {
	return findTorsocksBinary()
}

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
