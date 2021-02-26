package hosting

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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
	port    int
	address string
	cert    []byte
	running bool
	server  *http.Server
}

const certServerPort = 8181

func newCertificateServer(dir string) (*webserver, error) {
	certFile := filepath.Join(dir, "cert.pem")
	if !fileExists(certFile) {
		return nil, errors.New("the certificate file do not exists")
	}

	cert, err := ioutil.ReadFile(filepath.Clean(certFile))
	if err != nil {
		return nil, err
	}

	port := config.GetRandomPort()
	address := net.JoinHostPort(defaultHost(), strconv.Itoa(port))

	s := &webserver{
		port:    port,
		address: address,
		cert:    cert,
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
		IdleTimeout:  120 * time.Minute,
	}

	log.WithFields(log.Fields{
		"address": address,
		"dir":     dir,
	}).Debug("Creating Mumble certificate HTTP server")

	return s, nil
}

func (h *webserver) start(onFails func(error)) {
	if h.running {
		log.Error("Certificate HTTP server is already running")
		return
	}

	go func() {
		log.WithFields(log.Fields{
			"address": h.address,
		}).Debug("Starting Mumble certificate HTTP server")

		h.running = true

		err := h.server.ListenAndServe()
		if err != http.ErrServerClosed {
			if onFails != nil {
				onFails(err)
			}
		}

		h.running = false
	}()
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
	log.Debug("handleCertificateRequest(): serving certificate content")
	fmt.Fprint(w, string(h.cert))
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
