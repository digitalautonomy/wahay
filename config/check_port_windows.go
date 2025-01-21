package config

import (
	"net"
	"strconv"
	"time"
)

var (
	listen      = net.Listen
	dialTimeout = net.DialTimeout
)

// IsPortAvailable return a boolean indicating if a specific
// port is available to use
func IsPortAvailable(port int) bool {
	addr := net.JoinHostPort("", strconv.Itoa(port))
	conn, err := dialTimeout("tcp", addr, time.Second)

	if err == nil {
		conn.Close()
		return false
	}

	ln, err := listen("tcp", addr)
	if err != nil {
		return false
	}

	ln.Close()
	return true
}
