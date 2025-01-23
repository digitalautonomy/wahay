//go:build !windows
// +build !windows

package config

import "os/user"

// IsWindows returns true if this is running under windows
func IsWindows() bool {
	return false
}

// SystemDataDir points to the function that gets the data directory for this system
var (
	SystemDataDir   = XdgDataHome
	SystemConfigDir = XdgConfigHome
)

func localHome() string {
	u, e := user.Current()
	if e == nil {
		return u.HomeDir
	}
	return ""
}
