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

func (i *instance) exec(mainArg string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	// This executes the tor command, and the args which are both under control of the code
	/* #nosec G204 */
	cmd := exec.CommandContext(ctx, mainArg, args...)

	pathTorsocks, err := findLibTorsocks(i.pathTorsocks)
	if err != nil {
		cancelFunc()
		return nil, errors.New("error: libtorsocks.so was not found")
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

	if *config.Debug {
		cmd.Stdout = osf.Stdout()
		cmd.Stderr = osf.Stderr()
	}

	if err := execf.StartCommand(cmd); err != nil {
		cancelFunc()
		return nil, err
	}

	rc := &RunningCommand{
		Ctx:        ctx,
		Cmd:        cmd,
		CancelFunc: cancelFunc,
	}

	return rc, nil
}
