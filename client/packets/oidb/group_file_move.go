package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFileMoveReq(groupUin uint32, fileID string, parentFolder string, targetFolderID string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D6{
		Move: &oidb.OidbSvcTrpcTcp0X6D6Move{
			GroupUin:        groupUin,
			BusId:           102,
			FileId:          fileID,
			ParentDirectory: parentFolder,
			TargetDirectory: targetFolderID,
		},
	}
	return BuildOidbPacket(0x6D6, 5, body, false, true)
}

func ParseGroupFileMoveResp(data []byte) error {
	var resp oidb.OidbSvcTrpcTcp0X6D6Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return err
	}
	if resp.Move.RetCode != 0 {
		return errors.New(resp.Move.ClientWording)
	}
	return nil
}
