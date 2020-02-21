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

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

type webserver struct {
	sync.WaitGroup
	host    string
	port    int
	address string
	dir     string
	server  *http.Server
}

var initialized bool

func ensureCertificateServer(port int, dir string) (*webserver, error) {
	if !config.IsPortAvailable(port) {
		return nil, fmt.Errorf("the web server port is in use: %d", port)
	}

	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	httpServer := &http.Server{Addr: address}

	s := &webserver{
		host:    "127.0.0.1",
		port:    port,
		address: address,
		server:  httpServer,
		dir:     dir,
	}

	log.WithFields(log.Fields{
		"address": address,
		"dir":     dir,
	}).Info("Creating Mumble certificate server")

	if !initialized {
		http.HandleFunc("/", s.handleCertificateRequest)
		initialized = true
	}

	return s, nil
}

func (s *webserver) start() {
	s.Add(1)

	go s.startToListen()

	s.Wait()
}

func (s *webserver) startToListen() {
	defer s.Done()

	go func() {
		log.WithFields(log.Fields{
			"address": s.address,
			"dir":     s.dir,
		}).Info("Starting Mumble certificate server directory")

		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("certificate web server start(): %v", err)
		}
	}()
}

func (s *webserver) stop() error {
	return s.server.Shutdown(context.TODO())
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
