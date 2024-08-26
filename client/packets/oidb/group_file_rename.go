package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFileRenameReq(groupUin uint32, fileID string, parentFolder string, newFileName string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D6{
		Rename: &oidb.OidbSvcTrpcTcp0X6D6Rename{
			GroupUin:     groupUin,
			BusId:        102,
			FileId:       fileID,
			ParentFolder: parentFolder,
			NewFileName:  newFileName,
		},
	}
	return BuildOidbPacket(0x6D6, 4, body, false, true)
}

func ParseGroupFileRenameResp(data []byte) error {
	var resp oidb.OidbSvcTrpcTcp0X6D6Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return err
	}
	if resp.Rename.RetCode != 0 {
		return errors.New(resp.Rename.ClientWording)
	}
	return nil
}
