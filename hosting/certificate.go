package hosting

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

type webserver struct {
	host    string
	port    int
	address string
	dir     string
	wg      *sync.WaitGroup
	server  *http.Server
}

func ensureCertificationServer(port int, dir string) (*webserver, error) {
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
		wg:      &sync.WaitGroup{},
	}

	log.WithFields(log.Fields{
		"address": address,
		"dir":     dir,
	}).Info("Creating Mumble certication server")

	http.HandleFunc("/", s.handleCertificationRequest)

	return s, nil
}

func (s *webserver) start() {
	go func() {
		defer s.wg.Done()

		log.WithFields(log.Fields{
			"address": s.address,
			"dir":     s.dir,
		}).Info("Starting Mumble certication server directory")

		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("certificate web server start(): %v", err)
		}
	}()

	s.wg.Wait()
}

func (s *webserver) stop() error {
	return s.server.Shutdown(context.TODO())
}

func (s *webserver) handleCertificationRequest(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{
		"dir": s.dir,
	}).Debug("handleCertificationRequest(): serving file directory")

	http.ServeFile(w, r, filepath.Join(s.dir, "cert.pem"))
}
