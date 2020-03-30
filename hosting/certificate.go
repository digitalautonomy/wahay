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

type webserver struct {
	sync.WaitGroup
	address string
	dir     string
	running bool
	server  *http.Server
}

const certServerPort = 8181

func ensureCertificateServer(dir string) (*webserver, error) {
	if !config.IsPortAvailable(certServerPort) {
		return nil, fmt.Errorf("the web server port is in use: %d", certServerPort)
	}

	address := net.JoinHostPort(defaultHost, strconv.Itoa(certServerPort))

	s := &webserver{
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
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  4 * time.Minute,
	}

	log.WithFields(log.Fields{
		"address": address,
		"dir":     dir,
	}).Debug("Creating Mumble certificate HTTP server")

	return s, nil
}

func (h *webserver) host() string {
	return defaultHost
}

func (h *webserver) port() int {
	return certServerPort
}

func (h *webserver) start() {
	if h.running {
		log.Error("Certificate HTTP server is already running")
		return
	}

	go func() {
		log.WithFields(log.Fields{
			"address": h.address,
			"dir":     h.dir,
		}).Debug("Starting Mumble certificate HTTP server directory")

		// TODO[OB] - There's no way for the caller to know that this failed...
		err := h.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Fatal Mumble certificate HTTP server error: %v", err)
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
	}).Debug("handleCertificateRequest(): serving certificate file")

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
