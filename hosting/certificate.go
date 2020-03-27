package hosting

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

// TODO[OB] - why do we have the host when it's hard coded?
// TODO[OB] - why do we store the port? it doesn't seem to be used

type webserver struct {
	sync.WaitGroup
	host    string
	port    int
	address string
	dir     string
	running bool
	server  *http.Server
}

func ensureCertificateServer(port int, dir string) (*webserver, error) {
	if !config.IsPortAvailable(port) {
		return nil, fmt.Errorf("the web server port is in use: %d", port)
	}

	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	s := &webserver{
		host:    "127.0.0.1",
		port:    port,
		address: address,
		dir:     dir,
	}

	h := http.NewServeMux()
	h.HandleFunc("/", s.handleCertificateRequest)

	s.server = &http.Server{
		Addr:    address,
		Handler: h,

		// Set sensible timeouts, in case no reverse proxy is in front of Grumble.
		// Non-conforming (or malicious) clients may otherwise block indefinitely and cause
		// file descriptors (or handles, depending on your OS) to leak and/or be exhausted
		// TODO[OB] - I think these timeouts are too low. Remember that we are doing this over
		// Tor hidden services
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  2 * time.Minute,
	}

	// TODO[OB] - I'm not sure if it makes sense to log these as info
	log.WithFields(log.Fields{
		"address": address,
		"dir":     dir,
	}).Info("Creating Mumble certificate server")

	return s, nil
}

func (h *webserver) start() {
	if h.running {
		// TODO[OB] - why is this not an error?
		log.Warning("http server is already running")
		return
	}

	go func() {
		// TODO[OB] - not sure it makese sense to log as info
		log.WithFields(log.Fields{
			"address": h.address,
			"dir":     h.dir,
		}).Info("Starting Mumble certificate server directory")

		// TODO[OB] - There's no way for the caller to know that this failed...
		err := h.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Fatal HTTP server error: %v", err)
		}
	}()

	// TODO[OB] - There's a race condition here - the h.running can be set
	// before the server is listenting.
	h.running = true
}

func (h *webserver) stop() error {
	if !h.running {
		log.WithFields(log.Fields{
			"address": h.address,
		}).Debugf("stop(): http server not running")

		// we don't throw an error here because it's not a big deal
		// that the server is not running
		return nil
	}

	// TODO[OB] - I'm confused about this pattern
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	err := h.server.Shutdown(ctx)
	cancel()

	if err == context.DeadlineExceeded {
		log.Warning("Forcibly shutdown HTTP server while stopping")
	} else if err != nil {
		return err
	}

	h.running = false

	log.Info("HTTP server stopped")

	return nil
}

func (h *webserver) handleCertificateRequest(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"dir": h.dir,
	}).Debug("handleCertificateRequest(): serving cert file")

	// TODO[OB] - Why do we serve this from a file, and not from memory?
	if !fileExists(filepath.Join(h.dir, "cert.pem")) {
		http.Error(
			w,
			"The requested certificate is invalid",
			http.StatusInternalServerError,
		)
		return
	}

	http.ServeFile(w, r, filepath.Join(h.dir, "cert.pem"))
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
