package forwarder

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

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
	isRunning     bool
}

func NewForwarder(data hosting.MeetingData) *Forwarder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Forwarder{
		OnionAddr:     fmt.Sprintf("%s:%d", data.MeetingID, data.Port),
		LocalAddr:     "127.0.0.1",
		ListeningPort: assignPort(data),
		data:          data,
		ctx:           ctx,
		cancel:        cancel,
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
		log.Printf("Error parsing proxy URL: %v", err)
		return false
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		log.Printf("Error creating SOCKS5 dialer: %v", err)
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		Timeout:   5 * time.Second,
	}

	resp, err := client.Get("https://check.torproject.org/api/ip")
	if err != nil {
		log.Printf("Error reaching Tor check service: %v", err)
		return false
	}

	defer resp.Body.Close()
	return true
}

func (f *Forwarder) MonitorOnionService() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			log.Println("Stopping Onion service monitor...")
			return
		case <-ticker.C:
			if !f.CheckConnection() {
				log.Println("Onion service is unreachable; stopping forwarder.")
				f.StopForwarder()
				for {
					if f.CheckConnection() {
						log.Println("Onion service is reachable; starting forwarder.")
						go f.StartForwarder()
						break
					}
					time.Sleep(1 * time.Second)
				}
			} else if !f.isRunning {
				log.Println("Onion service is reachable; starting forwarder.")
				go f.StartForwarder()
			}
		}
	}
}

func (f *Forwarder) StartForwarder() {
	ctx, cancel := context.WithCancel(context.Background())
	f.ctx = ctx
	f.cancel = cancel
	listeningAddr := fmt.Sprintf("%s:%d", f.LocalAddr, f.ListeningPort)
	listener, err := net.Listen("tcp", listeningAddr)
	if err != nil {
		log.Printf("Failed to set up listener: %v\n", err)
	}

	f.l = listener
	defer func() {
		f.l.Close()
		f.l = nil
	}()

	log.Printf("TCP to SOCKS5 forwarder started on %s", listeningAddr)
	f.isRunning = true

	go f.MonitorOnionService()

	socks5Addr := fmt.Sprintf("%s:%d", f.LocalAddr, config.DefaultRoutePort)

	for {
		select {
		case <-f.ctx.Done():
			log.Println("Stopping forwarder...")
			f.isRunning = false
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
	if !f.isRunning {
		return
	}

	if f.cancel != nil {
		f.cancel()
	}

	f.l.Close()
	f.wg.Wait()
	log.Println("Forwarder stopped.")
	f.isRunning = false
}
