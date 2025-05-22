package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFileDeleteReq(groupUin uint32, fileID string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D6{
		Delete: &oidb.OidbSvcTrpcTcp0X6D6Delete{
			GroupUin: groupUin,
			BusId:    102,
			FileId:   fileID,
		},
	}
	return BuildOidbPacket(0x6D6, 3, body, false, true)
}

func ParseGroupFileDeleteResp(data []byte) error {
	var resp oidb.OidbSvcTrpcTcp0X6D6Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return err
	}
	if resp.Delete.RetCode != 0 {
		return errors.New(resp.Delete.ClientWording)
	}
	return nil
}
