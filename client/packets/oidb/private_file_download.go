package oidb

import (
	"fmt"

	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildPrivateFileDownloadReq(selfUid string, fileUUID string, fileHash string) (*OidbPacket, error) {
	body := &oidb2.OidbSvcTrpcTcp0XE37_1200{
		SubCommand: 1200,
		Field2:     1,
		Body: &oidb2.OidbSvcTrpcTcp0XE37_1200Body{
			ReceiverUid: selfUid,
			FileUuid:    fileUUID,
			Type:        2,
			FileHash:    fileHash,
			T2:          0,
		},
		Field101:   3,
		Field102:   103,
		Field200:   1,
		Field99999: []byte{0xc0, 0x85, 0x2c, 0x01},
	}
	return BuildOidbPacket(0xE37, 1200, body, false, false)
}

func ParsePrivateFileDownloadResp(data []byte) (string, error) {
	resp := new(oidb2.OidbSvcTrpcTcp0XE37_1200Response)
	if _, err := ParseOidbPacket(data, resp); err != nil {
		return "", err
	}
	url := fmt.Sprintf("http://%s:%d%s&isthumb=0",
		resp.Body.Result.Server,
		resp.Body.Result.Port,
		resp.Body.Result.Url,
	)
	return url, nil
}
