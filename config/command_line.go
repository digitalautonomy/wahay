package config

import (
	"flag"
)

// DefaultHost is where Tor is hosted
const DefaultHost = "127.0.0.1"

// DefaultRoutePort is the port Tor uses by default
const DefaultRoutePort = 9050

// DefaultControlPort is the port Tor uses by default for the control port
const DefaultControlPort = 9051

var (
	// TorHost contains the command line argument given for the Tor host
	TorHost = flag.String("tor-host", DefaultHost, "the host where Tor is listening")
	// TorPort contains the command line argument given for the Tor port
	TorPort = flag.Int("tor-port", DefaultControlPort, "the control port for Tor")
	// TorRoutePort contains the command line argument given for the Tor route port
	TorRoutePort = flag.Int("tor-route-port", DefaultRoutePort, "the route port for Tor")
	// TorControlPassword contains the command line argument given for the Tor control port password
	TorControlPassword = flag.String("tor-password", "", "the password for controlling Tor - can not be empty")
	// Debug contains the command line argument given for debugging
	Debug = flag.Bool("debug", false, "start Wahay in debugging mode")
	// Trace contains the command line argument given for debugging
	Trace = flag.Bool("trace", false, "start Wahay in tracing mode")
	// DebugFunctionCalls contains the command line argument given for debugging
	DebugFunctionCalls = flag.Bool("debug-function-calls", false, "trace function calls in logging")
)

// ProcessCommandLineArguments will parse the command line, check that
// required values are given and exit otherwise
func ProcessCommandLineArguments() {
	flag.Parse()
}
