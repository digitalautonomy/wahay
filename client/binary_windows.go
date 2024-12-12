package client

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	mumbleBundleLibsDir   = "lib"
	mumbleBundlePath      = "Mumble/client/mumble.exe"
	wahayMumbleBundlePath = "wahay/Mumble/client/mumble.exe"
)

var (
	execLookPath = exec.LookPath
	osGetenv     = os.Getenv
)

func searchBinaryInSystem() (*binary, error) {
	//Here we ignore the error because we handle the empty string returned.
	path, _ := execLookPath("mumble.exe")
	programFilesDir := osGetenv("PROGRAMFILES")
	programFilesX86Dir := osGetenv("PROGRAMFILES(X86)")
	dirs := []string{
		path,
		programFilesX86Dir,
		programFilesDir,
	}

	for _, d := range dirs {
		b := isThereAnAvailableBinary(d)
		if b != nil && b.isValid {
			return b, nil
		}
	}

	return nil, nil
}

func (b *binary) copyBinaryFilesToDir(destination string) error {
	err := filepath.Walk(filepath.Dir(b.path), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(b.path, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(destination, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info)
	})

	return err
}

func copyFile(src, dst string, info os.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer closeAndIgnore(srcFile)

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer closeAndIgnore(dstFile)

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode())
}
