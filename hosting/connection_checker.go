package hosting

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/digitalautonomy/wahay/config"
	log "github.com/sirupsen/logrus"
)

type checkService struct {
	port int
	l    net.Listener
	conn net.Conn
}

const checkConnPubPort = 12321

func newCheckConnectionService() (*checkService, error) {
	checkPort := config.GetRandomPort()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", checkPort))
	if err != nil {
		log.Errorf("Failed to start server on port %v: %v\n", checkPort, err)
		return nil, err
	}

	log.Infof("Check connection server listening on port: %v\n", checkPort)

	cs := &checkService{
		l:    listener,
		port: checkPort,
	}

	return cs, nil
}

func (cs *checkService) start() {
	go func() {
		for {
			conn, err := cs.l.Accept()
			if err != nil {
				log.Errorf("Error accepting connection: %v", err)
			}
			cs.conn = conn
			go cs.handleClient()
		}
	}()
}

func (cs *checkService) handleClient() {
	defer cs.conn.Close()

	for {
		message, err := cs.waitForClientMessage()
		if err != nil {
			log.Debug(err.Error())
			break
		}

		log.Debug(message)

		status, err := cs.sendConnectionConfirmation()
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		log.Debug(status)
	}
}

func (cs *checkService) waitForClientMessage() (string, error) {
	clientSignal, err := io.ReadAll(cs.conn)
	if err != nil && err != io.EOF {
		e := fmt.Sprintf("Error reading from connection: %s", err.Error())
		return "", errors.New(e)
	}

	if len(clientSignal) == 0 {
		return "", errors.New("received empty signal from client, waiting for new signal")
	}

	message := fmt.Sprintf("Received connection check signal from client: %s", string(clientSignal))

	return message, nil
}

func (cs *checkService) sendConnectionConfirmation() (string, error) {
	_, err := cs.conn.Write([]byte("OK\n"))
	if err != nil {
		e := fmt.Sprintf("Error writing response to connection: %s", err.Error())
		return "", errors.New(e)
	}
	return "OK signal send to client.", nil
}
