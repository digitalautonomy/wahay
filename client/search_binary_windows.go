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

	programFilesDir := os.Getenv("PROGRAMFILES")

	b := isThereAnAvailableBinary(programFilesDir)

	return b, nil
}
