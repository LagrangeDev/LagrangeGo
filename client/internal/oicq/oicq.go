package oicq

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/internal/oicq/oicq.go

import (
	"crypto/rand"
	goBinary "encoding/binary"

	tea "github.com/fumiama/gofastTEA"
	"github.com/pkg/errors"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

type Codec struct {
	ecdh      *session
	randomKey []byte

	WtSessionTicketKey []byte
}

func NewCodec(uin int64) *Codec {
	c := &Codec{
		ecdh:      newSession(),
		randomKey: make([]byte, 16),
	}
	_, _ = rand.Read(c.randomKey)
	c.ecdh.fetchPubKey(uin)
	return c
}

type EncryptionMethod byte

// nolint
const (
	EM_ECDH EncryptionMethod = iota
	EM_ST
)

type Message struct {
	Uin              uint32
	Command          uint16
	EncryptionMethod EncryptionMethod
	Body             []byte
}

func (c *Codec) Marshal(m *Message) []byte {
	w := binary.NewBuilder()

	w.WriteU8(0x02)
	w.WriteU16(0)    // len 占位
	w.WriteU16(8001) // version?
	w.WriteU16(m.Command)
	w.WriteU16(1)
	w.WriteU32(m.Uin)
	w.WriteU8(0x03)
	switch m.EncryptionMethod {
	case EM_ECDH:
		w.WriteU8(0x87)
	case EM_ST:
		w.WriteU8(0x45)
	}
	w.WriteU8(0)
	w.WriteU32(2)
	w.WriteU32(0)
	w.WriteU32(0)

	switch m.EncryptionMethod {
	case EM_ECDH:
		w.WriteU8(0x02)
		w.WriteU8(0x01)
		w.WriteBytes(c.randomKey)
		w.WriteU16(0x01_31)
		w.WriteU16(c.ecdh.SvrPublicKeyVer)
		w.WriteU16(uint16(len(c.ecdh.PublicKey)))
		w.WriteBytes(c.ecdh.PublicKey)
		w.EncryptAndWrite(c.ecdh.ShareKey, m.Body)

	case EM_ST:
		w.WriteU8(0x01)
		w.WriteU8(0x03)
		w.WriteBytes(c.randomKey)
		w.WriteU16(0x0102)
		w.WriteU16(0x0000)
		w.EncryptAndWrite(c.randomKey, m.Body)
	}
	w.WriteU8(0x03)

	data := w.ToBytes()
	buf := make([]byte, len(data))
	copy(buf, data)
	goBinary.BigEndian.PutUint16(buf[1:3], uint16(len(buf)))
	return buf
}

var (
	ErrUnknownFlag        = errors.New("unknown flag")
	ErrUnknownEncryptType = errors.New("unknown encrypt type")
)

func (c *Codec) Unmarshal(data []byte) (*Message, error) {
	reader := binary.NewReader(data)
	if flag := reader.ReadU8(); flag != 2 {
		return nil, ErrUnknownFlag
	}
	m := new(Message)
	reader.ReadU16() // len
	reader.ReadU16() // version?
	m.Command = reader.ReadU16()
	reader.ReadU16() // 1?
	m.Uin = reader.ReadU32()
	reader.ReadU8()
	encryptType := reader.ReadU8()
	reader.ReadU8()
	switch encryptType {
	case 0:
		d := reader.ReadBytes(reader.Len() - 1)
		defer func() {
			if pan := recover(); pan != nil {
				m.Body = tea.NewTeaCipher(c.randomKey).Decrypt(d)
			}
		}()
		m.Body = tea.NewTeaCipher(c.ecdh.ShareKey).Decrypt(d)
	case 3:
		d := reader.ReadBytes(reader.Len() - 1)
		m.Body = tea.NewTeaCipher(c.WtSessionTicketKey).Decrypt(d)
	default:
		return nil, ErrUnknownEncryptType
	}
	return m, nil
}

type TLV struct {
	Command uint16
	List    [][]byte
}

func (t *TLV) Marshal() []byte {
	w := binary.NewBuilder()

	w.WriteU16(t.Command)
	w.WriteU16(uint16(len(t.List)))
	for _, elem := range t.List {
		w.WriteBytes(elem)
	}

	return append([]byte(nil), w.ToBytes()...)
}

func (t *TLV) Append(b ...[]byte) {
	t.List = append(t.List, b...)
}
