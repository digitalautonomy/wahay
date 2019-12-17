package hosting

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/digitalautonomy/grumble/pkg/logtarget"
	"github.com/digitalautonomy/grumble/server"
)

// Things to do for Grumble:
//  - manage the data directory - we should create it and remove it at start and end
//  - we should divide it up into smaller functions that do one thing
//  - we don't have the blobstorage, maybe we don't need it
//  - we need to think about how to handle certificates
//  - we need to take a parameter that is the port to listen on
//  - we need a function to stop the server
//  - do we actually need to freeze the server? I think not
//  - clean up the log stuff a bit

func TestHosting() {
	var servers map[int64]*server.Server
	servers = make(map[int64]*server.Server)
	server.SetServers(servers)

	server.Args.DataDir = ".grumble-tmp"
	server.Args.LogPath = "grumble.log"

	dataDir, err := os.Open(server.Args.DataDir)
	if err != nil {
		log.Fatalf("Unable to open data directory (%v): %v", server.Args.DataDir, err)
		return
	}
	dataDir.Close()

	err = logtarget.Target.OpenFile(server.Args.LogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open log file (%v): %v", server.Args.LogPath, err)
		return
	}
	log.SetPrefix("[G] ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(&logtarget.Target)
	log.Printf("Grumble")
	log.Printf("Using data directory: %s", server.Args.DataDir)

	log.Printf("Generating 4096-bit RSA keypair for self-signed certificate...")

	certFn := filepath.Join(server.Args.DataDir, "cert.pem")
	keyFn := filepath.Join(server.Args.DataDir, "key.pem")
	err = server.GenerateSelfSignedCert(certFn, keyFn)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("Certificate output to %v", certFn)
	log.Printf("Private key output to %v", keyFn)
	// TODO:
	// - Set up Args to be correct
	// - set up blobstore

	serversDirPath := filepath.Join(server.Args.DataDir, "servers")
	err = os.Mkdir(serversDirPath, 0700)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Unable to create servers directory: %v", err)
	}

	s, err := server.NewServer(1)
	if err != nil {
		log.Fatalf("Couldn't create server: %s", err.Error())
	}
	servers[s.Id] = s
	s.Set("NoWebServer", "true")
	s.Set("Address", "127.0.0.1")
	//	s.Set("Port", "1234")

	os.Mkdir(filepath.Join(serversDirPath, fmt.Sprintf("%v", 1)), 0750)
	err = s.FreezeToFile()
	if err != nil {
		log.Fatalf("Unable to freeze newly created server to disk: %v", err.Error())
	}

	err = s.Start()
	if err != nil {
		log.Fatalf("Couldn't start server: %s", err.Error())
	}

	server.SignalHandler()
}
