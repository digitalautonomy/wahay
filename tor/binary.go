package tor

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

const libTorsocks = "libtorsocks.so"

var libPrefixes = []string{
	"",
	"/usr",
	"/usr/local",
}

var libSuffixes = []string{
	"",
	"/torsocks",
}

var (
	errInvalidCommand = errors.New("invalid command")
)

type binary struct {
	path     string
	env      []string
	isValid  bool
	isBundle bool
}

var (
	// ErrInvalidTorPath is an error to be trown where custom paths
	// to find the Tor binary are empty or don't exists
	ErrInvalidTorPath = errors.New("invalid Tor path")

	// ErrTorVersionNotCompatible is an error to be trown where some
	// Tor binary is found but the version is incompatible
	ErrTorVersionNotCompatible = errors.New("incompatible Tor version")

	// ErrInvalidConfiguredTorBinary is an error to be trown where the user
	// configure a custom path for Tor binary and it's no valid
	ErrInvalidConfiguredTorBinary = errors.New("invalid Tor binary user configured path")
)

func findTorBinary(conf *config.ApplicationConfig) (b *binary, err error) {
	functions := []func() (*binary, error){
		findTorBinaryInConfigPath(conf),
		findTorBinaryInDataDir,
		findTorBinaryInCurrentWorkingDir,
		findTorBinaryInWahayDir,
		findTorBinaryInSystem,
	}

	for _, cb := range functions {
		b, err = cb()
		if (b != nil && b.isValid) || err != nil {
			return
		}
	}

	// We only reach this point if we couldn't find a valid binary
	// It would be rude to return a binary that is invalid...
	b = nil
	err = ErrTorBinaryNotFound

	return
}

func findTorBinaryInConfigPath(conf *config.ApplicationConfig) func() (b *binary, fatalErr error) {
	return func() (*binary, error) {
		path := conf.GetPathTor()

		log.Debugf("findTorBinaryInConfigPath(%s)", path)

		// No configured path by the user to find Tor binary
		if len(path) == 0 {
			return nil, nil
		}

		// what should this one do?
		// if no Tor Path is configured, it just returns false
		// if a Tor path is configured, and it points to a valid Tor, we return it and true
		// if a Tor path is configured, but it is NOT valid, we will return an error
		//    This approach is FAIL CLOSED, we will not continue trying other Tors if the
		//    user has configured a specific Tor to use. This is conservative, and might limit
		//    functionality in some edge cases, but is significantly more secure
		b, err := isThereConfiguredTorBinary(conf.GetPathTor())
		if b == nil || err != nil {
			return nil, ErrInvalidConfiguredTorBinary
		}

		return b, nil
	}
}

func findTorBinaryInDataDir() (b *binary, fatalErr error) {
	paths := []string{
		"tor",
		"wahay/tor",
		"bin/wahay/tor",
	}

	for _, subdir := range paths {
		path := filepath.Join(config.XdgDataHome(), subdir)

		log.Debugf("findTorBinaryInDataDir(%s)", path)

		b, _ = isThereConfiguredTorBinary(path)
		if b != nil && b.isValid {
			return b, nil
		}
	}

	return nil, nil
}

func findTorBinaryInCurrentWorkingDir() (b *binary, fatalErr error) {
	log.Debugf("findTorBinaryInCurrentWorkingDir()")

	pathCWD, err := osf.Getwd()
	if err != nil {
		return nil, nil
	}

	paths := []string{
		"tor",
		"bin/tor",
	}

	for _, subdir := range paths {
		path := filepath.Join(pathCWD, subdir)

		b, _ = isThereConfiguredTorBinary(path)
		if b != nil && b.isValid {
			return b, nil
		}
	}

	return nil, nil
}

func findTorBinaryInWahayDir() (b *binary, fatalErr error) {
	abs, err := filepath.Abs(filepath.Dir(osf.Args()[0]))
	if err != nil {
		return nil, nil
	}

	path := filepath.Join(abs, "tor")

	log.Debugf("findTorBinaryInWahayDir(%s)", path)

	b, _ = isThereConfiguredTorBinary(path)

	return b, nil
}

func findTorBinaryInSystem() (b *binary, fatalErr error) {
	path, err := execf.LookPath("tor")
	if err != nil {
		return nil, nil
	}

	log.Debugf("findTorBinaryInSystem(%s)", path)

	b, errTorBinary := isThereConfiguredTorBinary(path)

	// Ensure we have torsocks available in the system
	if errTorBinary == nil {
		errTorsocks := findTorsocksBinary()
		if errTorsocks != nil {
			return b, errTorsocks
		}
	}

	return b, nil
}

func isThereConfiguredTorBinary(path string) (b *binary, err error) {
	if len(path) == 0 {
		return b, ErrInvalidTorPath
	}

	if !filesystemf.IsADirectory(path) {
		// We ommit the error here because it's ok while
		// we are checking multiple possible paths where
		// the Tor binary can be
		b, _ = getBinaryForPath(path)
		return
	}

	list := listPossibleTorBinary(path)

	if len(list) > 0 {
		for _, p := range list {
			b, _ = getBinaryForPath(p)
			if b.isValid {
				return b, nil
			}
		}
	}

	return
}

func getBinaryForPath(path string) (b *binary, err error) {
	b = &binary{
		path:     path,
		isBundle: false,
		isValid:  false,
		env:      []string{},
	}

	if checkIfBinaryIsBundled(b) {
		b.isBundle = true
		b.env = append(b.env, fmt.Sprintf("LD_LIBRARY_PATH=%s", filepath.Dir(path)))
	}

	b.isValid = isTorVersionCompatible(b)
	if !b.isValid {
		err = ErrTorVersionNotCompatible
	}

	return b, err
}

func checkIfBinaryIsBundled(b *binary) bool {
	libs := []string{
		"libcrypto*.so.*",
		"libevent*.so.*",
		"libssl*.so.*",
	}

	found := 0
	for _, l := range libs {
		matches, err := filepathf.Glob(filepath.Join(filepath.Dir(b.path), l))
		if err != nil {
			continue
		}

		if len(matches) != 0 {
			found++
		}
	}

	return found >= len(libs)
}

func isTorVersionCompatible(b *binary) bool {
	output, err := execTorCommand(b.path, []string{"--version"}, func(cmd *exec.Cmd) {
		if b.isBundle {
			cmd.Env = append(cmd.Env, b.env...)
		}
	})

	if len(output) == 0 || err != nil {
		return false
	}

	diff, err := compareVersions(extractVersionFrom(output), minSupportedVersion)
	if err != nil {
		return false
	}

	return diff >= 0
}

func execTorCommand(bin string, args []string, cm ModifyCommand) ([]byte, error) {
	output, err := execf.ExecWithModify(bin, args, cm)
	if len(output) == 0 || err != nil {
		return nil, errInvalidCommand
	}

	return output, nil
}

func extractVersionFrom(s []byte) string {
	r := regexp.MustCompile(`(\d+\.)(\d+\.)(\d+\.)(\d)`)
	result := r.FindStringSubmatch(string(s))

	if len(result) == 0 {
		return ""
	}

	return result[0]
}

func allLibDirs() []string {
	result := make([]string, 0)
	for _, l := range libDirs {
		for _, lp := range libPrefixes {
			for _, ls := range libSuffixes {
				result = append(result, filepath.Join(lp, l, ls))
			}
		}
	}
	return result
}

func findLibTorsocks(filePath string) (string, error) {
	// Search in user config path
	f := filepath.Join(filePath, libTorsocks)
	if filesystemf.FileExists(f) {
		return f, nil
	}

	// Search in local directories
	for _, ld := range allLibDirs() {
		f = filepath.Join(ld, libTorsocks)
		if filesystemf.FileExists(f) {
			return f, nil
		}
	}

	// Search in bundle path
	pathCWD, err := filepath.Abs(filepath.Dir(osf.Args()[0]))
	if err == nil {
		c := filepath.Join(pathCWD, "tor/")
		f = filepath.Join(c, libTorsocks)
		if filesystemf.FileExists(f) {
			return f, nil
		}
	}

	return "", errors.New("libtorsocks not found")
}

func listPossibleTorBinary(path string) []string {
	result := make([]string, 0)

	matches, _ := filepathf.Glob(filepath.Join(path, "tor*"))

	for _, match := range matches {
		filename := filepath.Base(match)

		if filename == "tor" {
			result = append(result, match)
		} else {
			diff, err := compareVersions(extractVersionFrom([]byte(filename)), minSupportedVersion)
			if err == nil && diff >= 0 {
				result = append(result, match)
			}
		}
	}

	return result
}

func (b *binary) start(configFile string) (*runningTor, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	// This is safe since we control both the path and the configFile argument - there is
	// no user input to these
	/* #nosec G204 */
	cmd := exec.CommandContext(ctx, b.path, "-f", configFile)

	if b.isBundle && len(b.env) > 0 {
		log.Debugf("Tor is bundled with environment variables: %s", b.env)
		cmd.Env = append(osf.Environ(), b.env...)
	}

	if err := execf.StartCommand(cmd); err != nil {
		cancelFunc()
		return nil, err
	}

	state := &runningTor{
		cmd:               cmd,
		ctx:               ctx,
		cancelFunc:        cancelFunc,
		finished:          false,
		finishedWithError: nil,
		finishChannel:     make(chan bool, 100),
	}

	return state, nil
}
