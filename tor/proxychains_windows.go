package tor

import (
	"context"
	"os/exec"

	"github.com/digitalautonomy/wahay/config"
	localExec "github.com/digitalautonomy/wahay/exec"
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

func (i *instance) exec(mumbleBinary string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	exitGroup, err := NewProcessExitGroup()
	if err != nil {
		cancelFunc()
		return nil, err
	}

	mainArg := "proxychains_win32_x64"
	args = append([]string{mumbleBinary}, args...)
	// This executes the proxychains command, and the args which are both under control of the code
	/* #nosec G204 */
	cmd := exec.CommandContext(ctx, mainArg, args...)
	localExec.HideCommandWindow(cmd)

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

	if err := exitGroup.AddProcess(cmd.Process); err != nil {
		cancelFunc()
		exitGroup.Dispose()
		cmd.Process.Kill()
		return nil, err
	}

	rc := &RunningCommand{
		Ctx:        ctx,
		Cmd:        cmd,
		CancelFunc: cancelFunc,
		ExitGroup:  exitGroup,
	}

	return rc, nil
}
