package client

import (
	"os"
	"os/exec"
)

const (
	mumbleBundleLibsDir   = "lib"
	mumbleBundlePath      = "Mumble/client/mumble.exe"
	wahayMumbleBundlePath = "wahay/Mumble/client/mumble.exe"
)

var execLookPath = exec.LookPath

func searchBinaryInSystem() (*binary, error) {
	//Here we ignore the error because we handle the empty string returned.
	path, _ := execLookPath("mumble.exe")
	programFilesDir := os.Getenv("PROGRAMFILES")
	programFilesX86Dir := os.Getenv("PROGRAMFILES(X86)")
	dirs := []string{
		programFilesX86Dir,
		programFilesDir,
		path,
	}

	for _, d := range dirs {
		b := isThereAnAvailableBinary(d)
		if b != nil && b.isValid {
			return b, nil
		}
	}

	return nil, nil
}
