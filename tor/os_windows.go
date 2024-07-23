package tor

import (
	log "github.com/sirupsen/logrus"
)

func searchProxyTool() error {
	return findProxychainsBinary()
}

func findProxychainsBinary() (fatalErr error) {
	return findProxychainsInSystem()
}

func findProxychainsInSystem() (fatalErr error) {
	path, err := execf.LookPath("proxychains_win32_x64")
	if err != nil {
		log.Errorf("Proxychains is not installed in your system: %s", err.Error())
		return ErrProxychainsNotInstalled
	}
	log.Debugf("findProxychainsInSystem(%s)", path)

	return nil
}
