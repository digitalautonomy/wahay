package config

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	mrand "math/rand"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/cubiest/jibberjabber"
	"golang.org/x/text/language"
)

// RandomString returns a string randomly generated
func RandomString(dest []byte) error {
	src := make([]byte, (len(dest)/2)+1)

	if _, err := io.ReadFull(rand.Reader, src); err != nil {
		return err
	}

	copy(dest, hex.EncodeToString(src))

	return nil
}

// localHome is a function to return the home directory on Unix-like operating systems
// Once support for Windows or other operating systems have been added, this
// should be extracted to a file protected by build tags.
func localHome() string {
	u, e := user.Current()
	if e == nil {
		return u.HomeDir
	}
	return ""
}

func home() string {
	e := os.Getenv("HOME")
	if e != "" {
		return e
	}
	return localHome()
}

// WithHome returns the given relative file/dir with the $HOME prepended
func WithHome(file string) string {
	return filepath.Join(home(), file)
}

func xdgOrWithHome(env, or string) string {
	if os.Getenv(env) == "" {
		return WithHome(or)
	}

	return os.Getenv(env)
}

type XdgConfigHomeFunc func() string

var XdgConfigHome XdgConfigHomeFunc = xdgConfigHome

func xdgConfigHome() string {
	return xdgOrWithHome("XDG_CONFIG_HOME", ".config")
}

// XdgDataHome returns the standardized XDG Data directory
func XdgDataHome() string {
	return xdgOrWithHome("XDG_DATA_HOME", ".local/share")
}

// IsPortAvailable return a boolean indicating if a specific
// port is available to use

var listen = net.Listen

func IsPortAvailable(port int) bool {
	ln, err := listen("tcp", net.JoinHostPort("", strconv.Itoa(port)))

	if err != nil {
		return false
	}

	return ln.Close() == nil
}

var randomInt31 = mrand.Int31n

// RandomPort returns a random port
func RandomPort() int {
	/* #nosec G404 */
	return 10000 + int(randomInt31(50000))
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
	return port > 0 && port < 65536
}

var detectLanguage = jibberjabber.DetectLanguageTag

// DetectLanguage determine the language used in the host computer
func DetectLanguage() language.Tag {
	tag, _ := detectLanguage()
	if tag == language.Und {
		tag = language.English
	}
	return tag
}
