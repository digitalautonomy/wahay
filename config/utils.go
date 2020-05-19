package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/cubiest/jibberjabber"
	"golang.org/x/text/language"
)

// ParseYes returns true if the string is any combination of yes
func ParseYes(input string) bool {
	switch strings.ToLower(input) {
	case "y", "yes":
		return true
	}

	return false
}

// RandomString returns a string randomly generated
func RandomString(dest []byte) error {
	src := make([]byte, len(dest))

	if _, err := io.ReadFull(rand.Reader, src); err != nil {
		return err
	}

	copy(dest, hex.EncodeToString(src))

	return nil
}

// WithHome returns the given relative file/dir with the $HOME prepended
func WithHome(file string) string {
	return filepath.Join(os.Getenv("HOME"), file)
}

func xdgOrWithHome(env, or string) string {
	x := os.Getenv(env)
	if x == "" {
		x = WithHome(or)
	}
	return x
}

// FindFileInLocations will check each path and if that file exists return the file name and true
func FindFileInLocations(places []string) (string, bool) {
	for _, p := range places {
		if FileExists(p) {
			return p, true
		}
	}
	return "", false
}

// XdgConfigHome returns the standardized XDG Configuration directory
func XdgConfigHome() string {
	return xdgOrWithHome("XDG_CONFIG_HOME", ".config")
}

// XdgCacheDir returns the standardized XDG Cache directory
func XdgCacheDir() string {
	return xdgOrWithHome("XDG_CACHE_HOME", ".cache")
}

// XdgDataHome returns the standardized XDG Data directory
func XdgDataHome() string {
	return xdgOrWithHome("XDG_DATA_HOME", ".local/share")
}

// XdgDataDirs returns the standardized XDG Data directory
func XdgDataDirs() []string {
	x := os.Getenv("XDG_DATA_DIRS")
	return strings.Split(x, ":")
}

// IsPortAvailable return a boolean indicatin if a specific
// port is available to use
func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return false
	}

	return ln.Close() == nil
}

// RandomPort returns a random port
func RandomPort() int {
	return 10000 + int(mrand.Int31n(50000))
}

// GetRandomPort returns an available random port
func GetRandomPort() int {
	port := RandomPort()
	for !IsPortAvailable(port) {
		port = RandomPort()
	}
	return port
}

// CheckPort returns boolean indicating if the port is valid or not
func CheckPort(port int) bool {
	if port <= 0 || port > 65535 {
		return false
	}
	return true
}

// DetectLanguage determine the language used in the host computer
func DetectLanguage() language.Tag {
	tag, _ := jibberjabber.DetectLanguageTag()
	if tag == language.Und {
		tag = language.English
	}
	return tag
}
