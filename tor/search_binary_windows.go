package tor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func findTorExecutable(dir string) (string, error) {
	var torExePath string

	possibleTorPaths := findPossibleTorPaths(dir)

	for _, rootDir := range possibleTorPaths {
		err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %s: %v\n", path, err)
				return nil
			}

			if d.IsDir() {
				return nil
			}

			if strings.EqualFold(d.Name(), "tor.exe") {
				torExePath = path
				return filepath.SkipDir
			}
			return nil
		})

		if torExePath != "" {
			return torExePath, nil
		}

		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("tor.exe not found in directory %s", dir)
}

func findPossibleTorPaths(parentDir string) []string {
	var torPaths []string

	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "tor") {
			torPaths = append(torPaths, filepath.Join(parentDir, entry.Name()))
		}
	}

	return torPaths
}

func isThereConfiguredTorBinary(path string) (b *binary, err error) {
	if len(path) == 0 {
		return b, ErrInvalidTorPath
	}

	if !filesystemf.IsADirectory(path) {
		b, err = getBinaryForPath(path)
		return
	}

	torExecutablePath, _ := findTorExecutable(path)

	b, err = getBinaryForPath(torExecutablePath)
	return
}

func findTorBinaryInSystem() (b *binary, fatalErr error) {
	//Here we ignore the error because we handle the empty string returned.
	path, _ := execf.LookPath("tor.exe")
	homeDir, _ := os.UserHomeDir()
	desktopDir := filepathf.Join(homeDir, "Desktop")
	oneDriveDesktopDir := filepathf.Join(homeDir, "OneDrive/Desktop")
	programFilesDir := osf.Getenv("PROGRAMFILES")
	programFilesX86Dir := osf.Getenv("PROGRAMFILES(X86)")

	dirs := []string{
		path,
		desktopDir,
		oneDriveDesktopDir,
		programFilesDir,
		programFilesX86Dir,
	}

	for _, d := range dirs {

		log.Debugf("findTorBinaryInSystem(%s)", d)

		b, _ = isThereConfiguredTorBinary(d)
		if b != nil && b.isValid {
			return b, nil
		}
	}

	return nil, nil
}
