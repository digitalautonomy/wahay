package tor

import (
	"context"
	"os/exec"

	"github.com/digitalautonomy/wahay/config"
	localExec "github.com/digitalautonomy/wahay/exec"
)

func (i *instance) exec(mumbleBinary string, args []string, pre ModifyCommand) (*RunningCommand, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	exitGroup, err := NewProcessExitGroup()
	if err != nil {
		cancelFunc()
		return nil, err
	}

	// This executes the mumble command, and the args which are both under control of the code
	/* #nosec G204 */
	cmd := exec.CommandContext(ctx, mumbleBinary, args...)
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
