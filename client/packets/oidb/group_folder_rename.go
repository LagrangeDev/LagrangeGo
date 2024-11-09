package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFolderRenameReq(groupUin uint32, folderID string, newFolderName string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D7{
		Rename: &oidb.OidbSvcTrpcTcp0X6D7Rename{
			GroupUin:      groupUin,
			FolderId:      folderID,
			NewFolderName: newFolderName,
		},
	}
	return BuildOidbPacket(0x6D7, 2, body, false, true)
}

func ParseGroupFolderRenameResp(data []byte) error {
	var resp oidb.OidbSvcTrpcTcp0X6D7Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return err
	}
	if resp.Rename.RetCode != 0 {
		return errors.New(resp.Rename.ClientWording)
	}
	return nil
}
