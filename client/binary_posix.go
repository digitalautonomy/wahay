//go:build !windows

package client

import (
	"io"
	"os"
	"os/exec"
)

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

func (b *binary) copyBinaryFilesToDir(destination string) error {
	var err error
	var srcfd *os.File

	if srcfd, err = os.Open(b.path); err != nil {
		return err
	}
	defer closeAndIgnore(srcfd)

	var dstfd *os.File

	if dstfd, err = os.Create(destination); err != nil {
		return err
	}
	defer closeAndIgnore(dstfd)

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}

	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(b.path); err != nil {
		return err
	}

	return os.Chmod(destination, srcinfo.Mode())
}
