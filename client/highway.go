package client

import (
	"bytes"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/highway"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/RomiChan/protobuf/proto"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type UpBlock struct {
	CommandId  int
	Uin        uint
	Sequence   uint
	FileSize   uint64
	Offset     uint64
	Ticket     []byte
	FileMd5    []byte
	Block      []byte
	ExtendInfo []byte
	Timestamp  uint64
}

func (c *QQClient) SendUpBlockAsync(block UpBlock, server string) bool {
	head := &highway.DataHighwayHead{
		Version:   1,
		Uin:       proto.Some(strconv.Itoa(int(block.Uin))),
		Command:   proto.Some("PicUp.DataUp"),
		Seq:       uint32(block.Sequence),
		AppId:     uint32(c.appInfo.AppID),
		DataFlag:  16,
		CommandId: uint32(block.CommandId),
	}
	md5 := utils.Md5Digest(block.Block)
	segHead := &highway.SegHead{
		Filesize:      block.FileSize,
		DataOffset:    block.Offset,
		DataLength:    uint32(len(block.Block)),
		ServiceTicket: block.Ticket,
		Md5:           md5,
		FileMd5:       block.FileMd5,
	}
	loginHead := &highway.LoginSigHead{
		Uint32LoginSigType: 8,
		BytesLoginSig:      c.sig.Tgt,
		AppId:              uint32(c.appInfo.AppID),
	}
	highwayHead := &highway.ReqDataHighwayHead{
		MsgBaseHead:        head,
		MsgSegHead:         segHead,
		BytesReqExtendInfo: block.ExtendInfo,
		Timestamp:          block.Timestamp,
		MsgLoginSigHead:    loginHead,
	}
	isEnd := block.Offset+uint64(len(block.Block)) == block.FileSize
	packet := &binary.Builder{}
	packet.WriteBytes(block.Block, false)
	payload, err := SendPacketAsync(highwayHead, packet, server, isEnd)
	if err != nil {
		return false
	}
	resphead, respbody, err := ParsePacket(payload)
	if err != nil {
		return false
	}
	networkLogger.Infof("Highway Block Result: %d | %d | %x | %v", resphead.ErrorCode, resphead.MsgSegHead.RetCode, resphead.BytesRspExtendInfo, respbody)
	return resphead.ErrorCode == 0
}

func ParsePacket(data []byte) (head *highway.RespDataHighwayHead, body *binary.Reader, err error) {
	reader := binary.NewReader(data)
	if reader.ReadBytes(1)[0] == 0x28 {
		headlength := reader.ReadU32()
		bodylength := reader.ReadU32()
		err = proto.Unmarshal(reader.ReadBytes(int(headlength)), &head)
		if err != nil {
			return nil, nil, err
		}
		err = proto.Unmarshal(reader.ReadBytes(int(bodylength)), body)
		if err != nil {
			return nil, nil, err
		}
		if reader.ReadBytes(1)[0] == 0x29 {
			return head, body, nil
		}
	}
	return nil, nil, err
}

func SendPacketAsync(packet *highway.ReqDataHighwayHead, buffer *binary.Builder, serverURL string, end bool) ([]byte, error) {
	marshal, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}
	writer := binary.NewBuilder(nil).WriteBytes([]byte{0x28}, false).WriteU32(uint32(len(marshal))).WriteU32(uint32(buffer.Len())).WriteBytes(buffer.Data(), false).WriteBytes([]byte{0x29}, false)
	return SendDataAsync(writer.Data(), serverURL, end)
}

func SendDataAsync(packet []byte, serverURL string, end bool) ([]byte, error) {
	// Parse server URL
	server, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	// Create request
	content := bytes.NewBuffer(packet)
	req, err := http.NewRequest("POST", server.String(), content)
	if err != nil {
		return nil, err
	}

	// Set headers
	if end {
		req.Header.Set("Connection", "close")
	} else {
		req.Header.Set("Connection", "keep-alive")
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read response data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
