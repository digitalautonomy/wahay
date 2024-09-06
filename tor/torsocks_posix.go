//go:build !windows

package tor

import (
	"context"
	"os/exec"

	"github.com/digitalautonomy/wahay/config"
)

func (i *instance) exec(mainArg string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	// This executes the mumble command, and the args which are both under control of the code
	/* #nosec G204 */
	cmd := exec.CommandContext(ctx, mainArg, args...)

	cmd.Env = osf.Environ()

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
