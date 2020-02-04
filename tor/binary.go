package tor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"

	"autonomia.digital/tonio/app/config"
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
		findTorBinaryInTonioDir,
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

		log.Printf("findTorBinaryInConfigPath(%s)", path)

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
		"tonio/tor",
		"bin/tonio/tor",
		"bin/tonio/tor/tor",
	}

	for _, subdir := range paths {
		path := filepath.Join(config.XdgDataHome(), subdir)

		log.Printf("findTorBinaryInDataDir(%s)", path)

		b, valid, err := isThereConfiguredTorBinary(path)
		if valid || err != nil {
			return b, valid, err
		}
	}

	return nil, false, nil
}

func findTorBinaryInCurrentWorkingDir() (b *binary, valid bool, err error) {
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

		log.Printf("findTorBinaryInCurrentWorkingDir(%s)", path)

		b, valid, err := isThereConfiguredTorBinary(path)
		if valid || err != nil {
			return b, valid, err
		}
	}

	return nil, false, nil
}

func findTorBinaryInTonioDir() (b *binary, valid bool, err error) {
	abs, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, false, nil
	}

	path := filepath.Join(abs, "tor/tor")

	log.Printf("findTorBinaryInTonioDir(%s)", path)

	return isThereConfiguredTorBinary(path)
}

func findTorBinaryInSystem() (b *binary, valid bool, err error) {
	path, err := exec.LookPath("tor")
	if err != nil {
		return nil, false, nil
	}

	log.Printf("findTorBinaryInSystem(%s)", path)

	return isThereConfiguredTorBinary(path)
}

func isThereConfiguredTorBinary(path string) (b *binary, valid bool, err error) {
	if len(path) == 0 {
		return nil, false, errors.New("no tor binary path defined")
	}

	// log.Printf("isThereConfiguredTorBinary(%s)", path)

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

// FindLibTorsocks returns the path of libtorsocks it exist
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
