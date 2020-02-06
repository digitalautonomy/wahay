package tor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

const libTorsocks = "libtorsocks.so"

var libDirs = []string{"/lib", "/lib64", "/lib/x86_64-linux-gnu", "/lib64/x86_64-linux-gnu"}
var libPrefixes = []string{"", "/usr", "/usr/local"}
var libSuffixes = []string{"", "/torsocks"}

var (
	errInvalidCommand = errors.New("invalid command")
)

type binary struct {
	path     string
	env      []string
	isValid  bool
	isBundle bool
}

func findTorBinary(conf *config.ApplicationConfig) (b *binary, valid bool, err error) {
	functions := []func() (*binary, bool, error){
		findTorBinaryInConfigPath(conf),
		findTorBinaryInDataDir,
		findTorBinaryInCurrentWorkingDir,
		findTorBinaryInWahayDir,
		findTorBinaryInSystem,
	}

	for _, cb := range functions {
		b, valid, err = cb()
		if valid || err != nil {
			return
		}
	}

	return nil, false, nil
}

func findTorBinaryInConfigPath(conf *config.ApplicationConfig) func() (*binary, bool, error) {
	return func() (*binary, bool, error) {
		path := conf.GetPathTor()

		log.Debugf("findTorBinaryInConfigPath(%s)", path)

		// No configured path by the user to find Tor binary
		if len(path) == 0 {
			return nil, false, nil
		}

		// what should this one do?
		// if no Tor Path is configured, it just returns false
		// if a Tor path is configured, and it points to a valid Tor, we return it and true
		// if a Tor path is configured, but it is NOT valid, we will return an error
		//    This approach is FAIL CLOSED, we will not continue trying other Tors if the
		//    user has configured a specific Tor to use. This is conservative, and might limit
		//    functionality in some edge cases, but is significantly more secure
		b, valid, _ := isThereConfiguredTorBinary(conf.GetPathTor())
		if !valid {
			return nil, false, errors.New("tor binary invalid user configured path")
		}

		return b, valid, nil
	}
}

func findTorBinaryInDataDir() (b *binary, valid bool, err error) {
	paths := []string{
		"tor",
		"wahay/tor",
		"bin/wahay/tor",
	}

	for _, subdir := range paths {
		path := filepath.Join(config.XdgDataHome(), subdir)

		log.Debugf("findTorBinaryInDataDir(%s)", path)

		b, valid, err := isThereConfiguredTorBinary(path)
		if valid || err != nil {
			return b, valid, err
		}
	}

	return nil, false, nil
}

func findTorBinaryInCurrentWorkingDir() (b *binary, valid bool, err error) {
	log.Debugf("findTorBinaryInCurrentWorkingDir()")

	pathCWD, err := os.Getwd()
	if err != nil {
		return nil, false, nil
	}

	paths := []string{
		"tor",
		"bin/tor",
	}

	for _, subdir := range paths {
		path := filepath.Join(pathCWD, subdir)

		b, valid, err := isThereConfiguredTorBinary(path)
		if valid || err != nil {
			return b, valid, err
		}
	}

	return nil, false, nil
}

func findTorBinaryInWahayDir() (b *binary, valid bool, err error) {
	abs, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, false, nil
	}

	path := filepath.Join(abs, "tor")

	log.Debugf("findTorBinaryInWahayDir(%s)", path)

	return isThereConfiguredTorBinary(path)
}

func findTorBinaryInSystem() (b *binary, valid bool, err error) {
	path, err := exec.LookPath("tor")
	if err != nil {
		return nil, false, nil
	}

	log.Debugf("findTorBinaryInSystem(%s)", path)

	return isThereConfiguredTorBinary(path)
}

func isThereConfiguredTorBinary(path string) (b *binary, valid bool, err error) {
	if len(path) == 0 {
		return nil, false, errors.New("no tor binary path defined")
	}

	if isADirectory(path) {
		list := listPossibleTorBinary(path)

		if len(list) > 0 {
			for _, p := range list {
				b, valid, err := isThereConfiguredTorBinaryHelper(p)
				if valid || err != nil {
					return b, valid, err
				}
			}
		}

		return nil, false, nil
	}

	return isThereConfiguredTorBinaryHelper(path)
}

func isThereConfiguredTorBinaryHelper(path string) (b *binary, valid bool, err error) {
	b = &binary{
		path:     path,
		isValid:  false,
		isBundle: false,
		env:      []string{},
	}

	if checkTorIsABundle(b) {
		b.isBundle = true
		b.env = append(b.env, fmt.Sprintf("LD_LIBRARY_PATH=%s", filepath.Dir(path)))
	}

	if checkTorVersionCompatibility(b) {
		b.isValid = true
	}

	return b, b.isValid, nil
}

func checkTorIsABundle(b *binary) bool {
	libs := []string{
		"libcrypto*.so.*",
		"libevent*.so.*",
		"libssl*.so.*",
	}

	found := 0
	for _, l := range libs {
		matches, err := filepath.Glob(filepath.Join(filepath.Dir(b.path), l))
		if err != nil {
			continue
		}

		if len(matches) != 0 {
			found++
		}
	}

	return found >= len(libs)
}

func checkTorVersionCompatibility(b *binary) bool {
	output, err := execTorCommand(b.path, []string{"--version"}, func(cmd *exec.Cmd) {
		if b.isBundle {
			cmd.Env = append(cmd.Env, b.env...)
		}
	})

	if len(output) == 0 || err != nil {
		return false
	}

	diff, err := compareVersions(extractVersionFrom(output), MinSupportedVersion)
	if err != nil {
		return false
	}

	return diff >= 0
}

func execTorCommand(bin string, args []string, cm ModifyCommand) ([]byte, error) {
	cmd := exec.Command(bin, args...)

	if cm != nil {
		cm(cmd)
	}

	output, err := cmd.Output()
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

// FindLibTorsocks returns the path of libtorsocks if exist
func FindLibTorsocks(filePath string) (string, error) {
	//Search in user config path
	f := filepath.Join(filePath, libTorsocks)
	if config.FileExists(f) {
		return f, nil
	}

	//Search in local directories
	for _, ld := range allLibDirs() {
		f = filepath.Join(ld, libTorsocks)
		if config.FileExists(f) {
			return f, nil
		}
	}

	//Search in bundle path
	pathCWD, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		c := filepath.Join(pathCWD, "tor/")
		f = filepath.Join(c, libTorsocks)
		if config.FileExists(f) {
			return f, nil
		}
	}

	return "", errors.New("libtorsocks not found")
}

func isADirectory(path string) bool {
	dir, err := os.Stat(path)
	if err != nil {
		return false
	}

	return dir.IsDir()
}

func listPossibleTorBinary(path string) []string {
	result := make([]string, 0)

	matches, _ := filepath.Glob(filepath.Join(path, "tor*"))

	for _, match := range matches {
		filename := filepath.Base(match)

		if filename == "tor" {
			result = append(result, match)
		} else {
			diff, err := compareVersions(extractVersionFrom([]byte(filename)), MinSupportedVersion)
			if err == nil && diff >= 0 {
				result = append(result, match)
			}
		}
	}

	return result
}

func (b *binary) start(configFile string) (*runningTor, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, b.path, "-f", configFile)

	if b.isBundle && len(b.env) > 0 {
		log.Debugf("Tor is bundled with environment variables: %s", b.env)
		cmd.Env = append(os.Environ(), b.env...)
	}

	if err := cmd.Start(); err != nil {
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
