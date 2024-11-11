package network

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/internal/network/response.go

import (
	"fmt"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	tea "github.com/fumiama/gofastTEA"
	"github.com/pkg/errors"
)

type Response struct {
	Type        RequestType
	EncryptType EncryptType // EncryptType 对应旧 SSOHeader.Flag
	SequenceID  int32       // SequenceID 对应旧 SSOPacket.Seq
	Uin         int64       // Uin 对应旧 SSOHeader.Uin
	CommandName string      // CommandName 对应旧 SSOPacket.Cmd
	Body        []byte      // Body 对应旧 SSOPacket.Data

	Message string // Message 对应旧 SSOPacket.Extra

	// Request is the original request that obtained this response.
	// Request *Request
}

var (
	ErrSessionExpired       = errors.New("session expired")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrPacketDropped        = errors.New("packet dropped")
	ErrInvalidPacketType    = errors.New("invalid packet type")
)

func (t *Transport) ReadResponse(head []byte) (*Response, error) {
	resp := new(Response)
	r := binary.NewReader(head)
	resp.Type = RequestType(r.ReadI32())
	if resp.Type != RequestTypeLogin && resp.Type != RequestTypeSimple && resp.Type != RequestTypeNT {
		return resp, ErrInvalidPacketType
	}
	resp.EncryptType = EncryptType(r.ReadU8())
	_ = r.ReadU8() // 0x00?

	resp.Uin, _ = strconv.ParseInt(r.ReadStringWithLength("u32", true), 10, 64)
	body := r.ReadAll()
	switch resp.EncryptType {
	case EncryptTypeNoEncrypt:
		// nothing to do
	case EncryptTypeD2Key:
		body = tea.NewTeaCipher(t.Sig.D2Key).Decrypt(body)
	case EncryptTypeEmptyKey:
		emptyKey := make([]byte, 16)
		body = tea.NewTeaCipher(emptyKey).Decrypt(body)
	}
	err := t.readSSOFrame(resp, body)
	return resp, err
}

func (t *Transport) readSSOFrame(resp *Response, payload []byte) error {
	reader := binary.NewReader(payload)
	headLen := reader.ReadI32()
	if headLen < 4 || headLen-4 > int32(reader.Len()) {
		return errors.WithStack(ErrPacketDropped)
	}

	head := binary.NewReader(reader.ReadBytes(int(headLen) - 4))
	resp.SequenceID = head.ReadI32()
	retCode := head.ReadI32()
	resp.Message = head.ReadStringWithLength("u32", true)
	var err error
	switch retCode {
	case 0:
		// ok
	case -10001, -10008: // -10001正常缓存过期，-10003登录失效？
		err = ErrSessionExpired
		fallthrough
	case -10003:
		err = ErrAuthenticationFailed
	default:
		err = errors.Errorf("return code unsuccessful: %d", retCode)
	}
	if err != nil {
		return errors.Errorf("%s %s", err.Error(), resp.Message)
	}
	resp.CommandName = head.ReadStringWithLength("u32", true)
	if resp.CommandName == "Heartbeat.Alive" {
		return nil
	}
	head.SkipBytesWithLength("u32", true) // session id
	compressedFlag := head.ReadI32()

	bodyLen := reader.ReadI32() - 4
	body := reader.ReadAll()
	if bodyLen > 0 && bodyLen < int32(len(body)) {
		body = body[:bodyLen]
	}
	switch compressedFlag {
	case 0:
	case 1:
		body = binary.ZlibUncompress(body)
	case 8:
		if len(body) > 4 {
			body = body[4:]
		} else {
			body = nil
		}
	default:
		return fmt.Errorf("unsupported compress flag %d", compressedFlag)
	}
	resp.Body = body
	return nil
}
