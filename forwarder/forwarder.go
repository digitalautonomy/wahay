package forwarder

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/hosting"
	log "github.com/sirupsen/logrus"
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
	isRunning     bool
	isPaused      bool
	pauseLock     sync.Mutex
	pausing       *pausing
}

func NewForwarder(data hosting.MeetingData) *Forwarder {
	f := &Forwarder{
		OnionAddr:     fmt.Sprintf("%s:%d", data.MeetingID, data.Port),
		LocalAddr:     "127.0.0.1",
		ListeningPort: assignPort(data),
		data:          data,
		pausing:       newPausing(),
	}

	f.pausing.check = f.CheckConnection
	f.pausing.onPause = f.onPause
	f.pausing.onWake = f.onWake

	return f
}

func (f *Forwarder) onPause() {
	f.pauseLock.Lock()
	defer f.pauseLock.Unlock()

	if f.l != nil {
		err := f.l.Close()
		if err != nil {
			log.Errorf("Error closing listener: %v", err)
		}
	}

	f.isPaused = true
	log.Debug("Forwarder paused.")
}

func (f *Forwarder) onWake() {
	f.pauseLock.Lock()
	defer f.pauseLock.Unlock()

	if err := f.setupListener(); err != nil {
		log.Errorf("Failed to set up listener on wake: %v", err)
		return
	}

	f.isPaused = false
	log.Debug("Forwarder resumed.")

	go f.acceptConnections()

}

func (f *Forwarder) setupListener() error {
	listeningAddr := fmt.Sprintf("%s:%d", f.LocalAddr, f.ListeningPort)
	listener, err := net.Listen("tcp", listeningAddr)
	if err != nil {
		return fmt.Errorf("failed to set up listener: %w", err)
	}
	f.l = listener
	return nil
}

func (f *Forwarder) shutdownListener() {
	if f.l != nil {
		err := f.l.Close()
		if err != nil {
			log.Errorf("Error closing listener: %v", err)
		}
		f.l = nil
	}
}

func assignPort(data hosting.MeetingData) int {
	if !data.IsHost {
		return config.GetRandomPort()
	}
	return data.Port
}

func (f *Forwarder) HandleConnection(clientConn net.Conn, socks5Addr string) {
	dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
	if err != nil {
		log.Errorf("Failed to create SOCKS5 dialer: %v\n", err)
		return
	}

	serverConn, err := dialer.Dial("tcp", f.OnionAddr)
	if err != nil {
		log.Errorf("Failed to connect to Mumble server via SOCKS5: %v\n", err)
		return
	}

	tcpClientConn, _ := clientConn.(*net.TCPConn)
	tcpServerConn, _ := serverConn.(*net.TCPConn)

	f.forwardTraffic(tcpClientConn, tcpServerConn)
}

func (f *Forwarder) forwardTraffic(conn1, conn2 *net.TCPConn) {
	defer conn1.Close()
	defer conn2.Close()

	f.wg.Add(2)

	go func() {
		defer f.wg.Done()
		io.Copy(conn1, conn2)
		conn1.CloseWrite()
	}()
	go func() {
		defer f.wg.Done()
		io.Copy(conn2, conn1)
		conn2.CloseWrite()
	}()

	f.wg.Wait()
}

func (f *Forwarder) CheckConnection() bool {
	proxyURL, err := url.Parse("socks5://" + net.JoinHostPort(f.LocalAddr, strconv.Itoa(config.DefaultRoutePort)))
	if err != nil {
		log.Errorf("Error parsing proxy URL: %v", err)
		return false
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		log.Errorf("Error creating SOCKS5 dialer: %v", err)
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		Timeout:   5 * time.Second,
	}

	resp, err := client.Get("https://check.torproject.org/api/ip")
	if err != nil {
		log.Errorf("Error reaching Tor check service: %v", err)
		return false
	}

	defer resp.Body.Close()
	return true
}

func (f *Forwarder) StartForwarder() {
	ctx, cancel := context.WithCancel(context.Background())
	f.ctx = ctx
	f.cancel = cancel
	f.isRunning = true
	f.isPaused = false

	f.pausing.run()

	if err := f.setupListener(); err != nil {
		log.Errorf("Failed to set up listener in StartForwarder: %v", err)
		f.isRunning = false
		return
	}

	log.Debugf("TCP to SOCKS5 forwarder started on %s:%d", f.LocalAddr, f.ListeningPort)

	go f.acceptConnections()

	<-ctx.Done()

	log.Debug("Forwarder stopping...")
}

func (f *Forwarder) acceptConnections() {
	socks5Addr := fmt.Sprintf("%s:%d", f.LocalAddr, config.DefaultRoutePort)

	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			f.pauseLock.Lock()
			if f.isPaused {
				f.pauseLock.Unlock()
				return
			}
			f.pauseLock.Unlock()

			clientConn, err := f.l.Accept()
			if err != nil {
				if f.isPaused {
					return
				}
				log.Errorf("Failed to accept connection: %v", err)
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
	if !f.isRunning {
		return
	}

	if f.cancel != nil {
		f.cancel()
	}

	f.shutdownListener()

	f.wg.Wait()
	f.pausing.stop()
	log.Debug("Forwarder stopped.")
	f.isRunning = false
}

type pausing struct {
	interval time.Duration
	paused   bool
	stopC    chan bool
	onPause  func()
	onWake   func()
	check    func() bool
}

func newPausing() *pausing {
	return &pausing{
		interval: 10 * time.Second,
		paused:   false,
		stopC:    make(chan bool),
	}
}

func (p *pausing) stop() {
	p.stopC <- true
}

func (p *pausing) run() {
	go p.runCheck()
}

func (p *pausing) runCheck() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopC:
			return
		case <-ticker.C:
			result := p.check()
			if !result && !p.paused {
				p.pause()
			} else if result && p.paused {
				p.wake()
			}
		}
	}
}

func (p *pausing) pause() {
	p.paused = true
	p.onPause()
}

func (p *pausing) wake() {
	p.paused = false
	p.onWake()
}
