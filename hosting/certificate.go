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
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  2 * time.Minute,
	}

	log.WithFields(log.Fields{
		"address": address,
		"dir":     dir,
	}).Info("Creating Mumble certificate server")

	return s, nil
}

func (s *webserver) start() {
	if s.running {
		log.Warning("http server is already running")
		return
	}

	go func() {
		log.WithFields(log.Fields{
			"address": s.address,
			"dir":     s.dir,
		}).Info("Starting Mumble certificate server directory")

		err := s.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("Fatal HTTP server error: %v", err)
		}
	}()

	s.running = true
}

func (s *webserver) stop() error {
	if !s.running {
		log.WithFields(log.Fields{
			"address": s.address,
		}).Debugf("stop(): http server not running")

		// we don't throw an error here because it's not a big deal
		// that the server is not running
		return nil
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	err := s.server.Shutdown(ctx)
	cancel()

	if err == context.DeadlineExceeded {
		log.Warning("Forcibly shutdown HTTP server while stopping")
	} else if err != nil {
		return err
	}

	s.running = false

	log.Info("HTTP server stopped")

	return nil
}

func (s *webserver) handleCertificateRequest(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"dir": s.dir,
	}).Debug("handleCertificateRequest(): serving cert file")

	if !fileExists(filepath.Join(s.dir, "cert.pem")) {
		http.Error(
			w,
			"The requested certificate is invalid",
			http.StatusInternalServerError,
		)
		return
	}

	http.ServeFile(w, r, filepath.Join(s.dir, "cert.pem"))
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
