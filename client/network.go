package client

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var ErrConnectionClosed = errors.New("connection closed")

type TCPClient struct {
	lock sync.RWMutex

	addr    string
	conn    net.Conn
	timeout int

	connected bool
}

func NewTCPClient(addr string, timeout int) *TCPClient {
	return &TCPClient{
		addr:    addr,
		timeout: timeout,
	}
}

func (c *TCPClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.addr, time.Duration(c.timeout)*time.Second)
	if err != nil {
		return err
	}
	networkLogger.Infof("connected to %s", conn.RemoteAddr())
	c.lock.Lock()
	defer c.lock.Unlock()
	c.conn = conn
	c.connected = true
	return nil
}

func (c *TCPClient) Write(b []byte) error {
	if conn := c.getConn(); conn != nil {
		n, err := conn.Write(b)
		if err != nil {
			return ErrConnectionClosed
		}
		networkLogger.Tracef("tcp write %d bytes", n)
		return nil
	}
	return ErrConnectionClosed
}

func (c *TCPClient) ReadBytes(len int) ([]byte, error) {
	if conn := c.getConn(); conn != nil {
		buffer := make([]byte, len)
		_, err := io.ReadFull(conn, buffer)
		if err != nil {
			c.onDisconnetced()
			return nil, ErrConnectionClosed
		}
		return buffer, nil
	}
	return nil, ErrConnectionClosed
}

func (c *TCPClient) onDisconnetced() {
	c.Close()
}
func (c *TCPClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		networkLogger.Error("tcp closed")
		c.conn = nil
		c.connected = false
	}
}

func (c *TCPClient) getConn() net.Conn {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.conn
}

func (c *TCPClient) IsClosed() bool {
	return !c.connected
}
