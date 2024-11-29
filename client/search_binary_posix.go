//go:build !windows

package client

import "os/exec"

const (
	mumbleBundleLibsDir   = "lib"
	mumbleBundlePath      = "mumble/mumble"
	wahayMumbleBundlePath = "wahay/mumble/mumble"
)

var execLookPath = exec.LookPath

func searchBinaryInSystem() (*binary, error) {
	path, err := execLookPath("mumble")
	if err != nil {
		return nil, nil
	}

	b := isThereAnAvailableBinary(path)

	return b, nil
}
