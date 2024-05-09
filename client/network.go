package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/internal/network/conn.go

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
	conn net.Conn
}

func (c *TCPClient) Connect(addr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	networkLogger.Infof("connected to %s", conn.RemoteAddr())
	c.lock.Lock()
	defer c.lock.Unlock()
	c.conn = conn
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
	}
}

func (c *TCPClient) getConn() net.Conn {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.conn
}

func (c *TCPClient) IsClosed() bool {
	return c.conn == nil
}
