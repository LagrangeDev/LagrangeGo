package binary

// from https://github.com/Mrs4s/MiraiGo/blob/master/binary/reader.go

import (
	"encoding/binary"
	"io"
	"net"
)

type NetworkReader struct {
	conn net.Conn
}

func NewNetworkReader(conn net.Conn) *NetworkReader {
	return &NetworkReader{conn: conn}
}

func (r *NetworkReader) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	n, err := r.conn.Read(buf)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return r.ReadByte()
	}
	return buf[0], nil
}

func (r *NetworkReader) ReadBytes(len int) ([]byte, error) {
	buf := make([]byte, len)
	_, err := io.ReadFull(r.conn, buf)
	return buf, err
}

func (r *NetworkReader) ReadInt32() (int32, error) {
	b := make([]byte, 4)
	_, err := r.conn.Read(b)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(b)), nil
}
