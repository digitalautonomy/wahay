package tor

import (
	"context"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

type ProcessExitGroup windows.Handle

// RunningCommand is a representation of a torify command
type RunningCommand struct {
	Ctx        context.Context
	Cmd        *exec.Cmd
	CancelFunc context.CancelFunc
	ExitGroup  ProcessExitGroup
}

func (s *service) Close() {
	s.rc.CancelFunc()
}

type process struct {
	Pid    int
	Handle uintptr
}

func NewProcessExitGroup() (ProcessExitGroup, error) {
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, err
	}

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(
		handle,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info))); err != nil {
		return 0, err
	}

	return ProcessExitGroup(handle), nil
}

func (g ProcessExitGroup) Dispose() error {
	return windows.CloseHandle(windows.Handle(g))
}

func (g ProcessExitGroup) AddProcess(p *os.Process) error {
	return windows.AssignProcessToJobObject(
		windows.Handle(g),
		windows.Handle((*process)(unsafe.Pointer(p)).Handle))
}
