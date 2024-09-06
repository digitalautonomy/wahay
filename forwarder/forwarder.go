package forwarder

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"sync"

	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/hosting"
	"golang.org/x/net/proxy"
)

type Forwarder struct {
	ListeningPort int
	LocalAddr     string
	OnionAddr     string
	data          hosting.MeetingData
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	l             net.Listener
}

func NewForwarder(data hosting.MeetingData) *Forwarder {
	ctx, cancel := context.WithCancel(context.Background()) // Create a context with cancel
	return &Forwarder{
		OnionAddr:     fmt.Sprintf("%s:%d", data.MeetingID, data.Port),
		LocalAddr:     "127.0.0.1",
		ListeningPort: 3000,
		data:          data,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (f *Forwarder) HandleConnection(clientConn net.Conn, socks5Addr string) {
	dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
	if err != nil {
		log.Printf("Failed to create SOCKS5 dialer: %v\n", err)
		return
	}

	serverConn, err := dialer.Dial("tcp", f.OnionAddr)
	if err != nil {
		log.Printf("Failed to connect to Mumble server via SOCKS5: %v\n", err)
		return
	}

	tcpClientConn, _ := clientConn.(*net.TCPConn)
	tcpServerConn, _ := serverConn.(*net.TCPConn)

	go f.forwardTraffic(tcpClientConn, tcpServerConn)
}

func (f *Forwarder) forwardTraffic(conn1, conn2 *net.TCPConn) {
	defer conn1.Close()
	defer conn2.Close()

	f.wg.Add(2)

	go func() {
		defer f.wg.Done()
		io.Copy(conn1, conn2)
		// Signal peer that no more data is coming.
		conn1.CloseWrite()
	}()
	go func() {
		defer f.wg.Done()
		io.Copy(conn2, conn1)
		// Signal peer that no more data is coming.
		conn2.CloseWrite()
	}()

	f.wg.Wait()
}

func (f *Forwarder) StartForwarder() {
	listeningAddr := fmt.Sprintf("%s:%d", f.LocalAddr, f.ListeningPort)
	listener, err := net.Listen("tcp", listeningAddr)
	if err != nil {
		log.Fatalf("Failed to set up listener: %v\n", err)
	}

	f.l = listener
	defer f.l.Close()

	log.Println("TCP to SOCKS5 forwarder is running...")

	socks5Addr := fmt.Sprintf("%s:%d", f.LocalAddr, config.DefaultRoutePort)

	for {
		select {
		case <-f.ctx.Done(): // Stop the forwarder when context is canceled
			log.Println("Stopping forwarder...")
			return
		default:
			clientConn, err := f.l.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v\n", err)
				continue
			}
			go f.HandleConnection(clientConn, socks5Addr)
		}
	}
}

func (f *Forwarder) GenerateURL() string {
	u := url.URL{
		Scheme: "mumble",
		User:   url.UserPassword(f.data.Username, f.data.Password),
		Host:   fmt.Sprintf("%s:%d", f.LocalAddr, f.ListeningPort),
	}

	return u.String()
}

func (f *Forwarder) StopForwarder() {
	if f.cancel != nil {
		f.cancel()
	}

	f.l.Close()

	f.wg.Wait()
	log.Println("Forwarder stopped.")
}
