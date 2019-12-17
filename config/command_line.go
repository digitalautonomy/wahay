package config

import (
	"flag"
	"fmt"
	"os"
)

var (
	// TorHost contains the command line argument given for the Tor host
	TorHost = flag.String("tor-host", "127.0.0.1", "the host where Tor is listening")
	// TorPort contains the command line argument given for the Tor port
	TorPort = flag.String("tor-port", "9051", "the control port for Tor")
	// TorControlPassword contains the command line argument given for the Tor control port password
	TorControlPassword = flag.String("tor-password", "", "the password for controlling Tor - can not be empty")
)

// ProcessCommandLineArguments will parse the command line, check that
// required values are given and exit otherwise
func ProcessCommandLineArguments() {
	flag.Parse()
	if *TorControlPassword == "" {
		fmt.Printf("For now, you have to supply a password for the Tor control port\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
