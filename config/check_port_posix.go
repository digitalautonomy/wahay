//go:build !windows
// +build !windows

package config

import (
	"net"
	"strconv"
)

var (
	listen      = net.Listen
	dialTimeout = net.DialTimeout
)

// IsPortAvailable return a boolean indicating if a specific
// port is available to use
func IsPortAvailable(port int) bool {
	ln, err := listen("tcp", net.JoinHostPort("", strconv.Itoa(port)))

	if err != nil {
		return false
	}

	return ln.Close() == nil
}
