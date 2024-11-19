package forwarder

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

const checkConnectionPort = 12321

type checker struct {
	conn net.Conn
}

func (f *Forwarder) CheckConnection() bool {
	err := f.connectToCheckerService()
	if err != nil {
		return false
	}
	defer f.checker.conn.Close()

	message := "Testing connection\n"
	_, err = f.checker.conn.Write([]byte(message))
	if err != nil {
		log.Errorf("Writing failed. Error: %v", err.Error())
		return false
	}
	log.Debug("Message sent")

	reader := bufio.NewReader(f.checker.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Debugf("Error while reading from connection: %s", err.Error())
		return false
	}

	log.Debugf("Server responds with: %v", response)

	if response != "OK\n" {
		log.Error("Connection lost or server not responding")
		return false
	}

	log.Debug("OK signal received:", response)
	return true
}

func (f *Forwarder) connectToCheckerService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := f.dialWithContext(ctx, "tcp", fmt.Sprintf("%s:%d", f.OnionAddr, checkConnectionPort))
	if err != nil {
		log.Debugf("Disconnected (no net or service unavailable): %v", err)
		return err
	}

	f.checker.conn = conn
	log.Debug("Connected to check service.")
	return nil
}

func (f *Forwarder) dialWithContext(ctx context.Context, network, address string) (net.Conn, error) {
	type result struct {
		conn net.Conn
		err  error
	}

	resultChan := make(chan result, 1)

	go func() {
		conn, err := f.dialer.Dial(network, address)
		resultChan <- result{conn: conn, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultChan:
		return res.conn, res.err
	}
}
