package oidb

import (
	"errors"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFolderCreateReq(groupUin uint32, targetDirectory string, folderName string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D7{
		Create: &oidb.OidbSvcTrpcTcp0X6D7Create{
			GroupUin:        groupUin,
			TargetDirectory: targetDirectory,
			FolderName:      folderName,
		},
	}
	return BuildOidbPacket(0x6D7, 0, body, false, true)
}

func ParseGroupFolderCreateResp(data []byte) error {
	var resp oidb.OidbSvcTrpcTcp0X6D7Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return err
	}
	if resp.Create.RetCode != 0 {
		return errors.New(resp.Create.ClientWording)
	}
	return nil
}
