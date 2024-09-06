package forwarder

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/digitalautonomy/wahay/config"
	"golang.org/x/net/proxy"
)

func HandleConnection(clientConn net.Conn, socks5Addr, onionAddr string) {
	dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
	if err != nil {
		log.Printf("Failed to create SOCKS5 dialer: %v\n", err)
		return
	}

	serverConn, err := dialer.Dial("tcp", onionAddr)
	if err != nil {
		log.Printf("Failed to connect to Mumble server via SOCKS5: %v\n", err)
		return
	}

	tcpClientConn, _ := clientConn.(*net.TCPConn)
	tcpServerConn, _ := serverConn.(*net.TCPConn)

	go forwardTraffic(tcpClientConn, tcpServerConn)
}

func forwardTraffic(conn1, conn2 *net.TCPConn) {
	defer conn1.Close()
	defer conn2.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(conn1, conn2)
		// Signal peer that no more data is coming.
		conn1.CloseWrite()
	}()
	go func() {
		defer wg.Done()
		io.Copy(conn2, conn1)
		// Signal peer that no more data is coming.
		conn2.CloseWrite()
	}()

	wg.Wait()
}

func StartForwarder(onionAddr string) {
	listener, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		log.Fatalf("Failed to set up listener: %v\n", err)
	}
	defer listener.Close()

	log.Println("TCP to SOCKS5 forwarder is running...")

	socks5Addr := fmt.Sprintf("127.0.0.1:%v", config.DefaultRoutePort)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go HandleConnection(clientConn, socks5Addr, onionAddr)
	}
}
