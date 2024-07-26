//go:build !windows

package tor

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/digitalautonomy/wahay/config"
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

func setupProxyToolEnvironment(i *instance, cmd *exec.Cmd, cancelFunc context.CancelFunc, pre ModifyCommand) error {
	pathTorsocks, err := findLibTorsocks(i.pathTorsocks)
	if err != nil {
		cancelFunc()
		return errors.New("error: libtorsocks.so was not found")
	}

	pwd := [32]byte{}
	_ = config.RandomString(pwd[:])

	cmd.Env = osf.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("LD_PRELOAD=%s", pathTorsocks))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TORSOCKS_PASSWORD=%s", string(pwd[:])))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TORSOCKS_TOR_ADDRESS=%s", i.controlHost))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TORSOCKS_TOR_PORT=%d", i.socksPort))

	if pre != nil {
		pre(cmd)
	}

	return nil
}
