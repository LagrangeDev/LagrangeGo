package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFolderDeleteReq(groupUin uint32, folderID string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D7{
		Delete: &oidb.OidbSvcTrpcTcp0X6D7Delete{
			GroupUin: groupUin,
			FolderId: folderID,
		},
	}
	return BuildOidbPacket(0x6D7, 1, body, false, true)
}

func ParseGroupFolderDeleteResp(data []byte) error {
	var resp oidb.OidbSvcTrpcTcp0X6D7Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return err
	}
	if resp.Delete.RetCode != 0 {
		return errors.New(resp.Delete.ClientWording)
	}
	return nil
}
