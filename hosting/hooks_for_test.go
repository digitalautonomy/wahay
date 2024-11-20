package hosting

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type hostingSuite struct{}

var _ = Suite(&hostingSuite{})

func noErrHelper() error {
	return nil
}
func errHelper() error {
	return errors.New("error2")
}

func mockCheckService() *checkService {
	listener := newMockListener()
	conn := newMockConn()

	listener.connChan <- conn

	return &checkService{
		l:    listener,
		conn: conn,
		port: 12345,
	}
}

type mockListener struct {
	connChan chan net.Conn
}

func newMockListener() *mockListener {
	return &mockListener{
		connChan: make(chan net.Conn, 1),
	}
}

func (m *mockListener) Accept() (net.Conn, error) {
	conn, ok := <-m.connChan
	if !ok {
		return nil, fmt.Errorf("mock listener closed")
	}
	return conn, nil
}

func (m *mockListener) Close() error {
	close(m.connChan)
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

type mockConn struct {
	readBuffer  *bytes.Buffer
	writeBuffer *bytes.Buffer
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuffer:  &bytes.Buffer{},
		writeBuffer: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return m.readBuffer.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeBuffer.Write(b)
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}
func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return nil
}
