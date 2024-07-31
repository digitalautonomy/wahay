//go:build !windows

package exec

import (
	"os/exec"
)

func HideCommandWindow(cmd *exec.Cmd) {}
