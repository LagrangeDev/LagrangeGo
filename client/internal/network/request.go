package network

import "github.com/LagrangeDev/LagrangeGo/client/sign"

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/internal/network/request.go

type RequestType uint32

const (
	RequestTypeLogin  = 0x0A
	RequestTypeSimple = 0x0B
	RequestTypeNT     = 0x0C
)

type EncryptType uint32

const (
	EncryptTypeNoEncrypt EncryptType = iota // 0x00
	EncryptTypeD2Key                        // 0x01
	EncryptTypeEmptyKey                     // 0x02
)

type Request struct {
	SequenceID  uint32
	Uin         int64
	Sign        *sign.Response
	CommandName string
	Body        []byte
}
