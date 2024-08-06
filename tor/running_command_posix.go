//go:build !windows

package tor

import (
	"context"
	"os/exec"
)

// RunningCommand is a representation of a torify command
type RunningCommand struct {
	Ctx        context.Context
	Cmd        *exec.Cmd
	CancelFunc context.CancelFunc
}

func (s *service) Close() {
	s.rc.CancelFunc()
}
