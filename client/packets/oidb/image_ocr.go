package oidb

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildImageOcrRequestPacket(url string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XE07_0{
		Version:  1,
		Client:   0,
		Entrance: 1,
		OcrReqBody: &oidb.OcrReqBody{
			ImageUrl:              url,
			OriginMd5:             "",
			AfterCompressMd5:      "",
			AfterCompressFileSize: "",
			AfterCompressWeight:   "",
			AfterCompressHeight:   "",
			IsCut:                 false,
		},
	}
	return BuildOidbPacket(0xE07, 0, body, false, true)
}

type (
	OcrResponse struct {
		Texts    []*TextDetection `json:"texts"`
		Language string           `json:"language"`
	}
	TextDetection struct {
		Text        string        `json:"text"`
		Confidence  int32         `json:"confidence"`
		Coordinates []*Coordinate `json:"coordinates"`
	}
	Coordinate struct {
		X int32 `json:"x"`
		Y int32 `json:"y"`
	}
)

func ParseImageOcrResp(data []byte) (*OcrResponse, error) {
	var rsp oidb.OidbSvcTrpcTcp0XE07_0_Response
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return nil, err
	}
	if rsp.Wording != "" {
		if strings.Contains(rsp.Wording, "服务忙") {
			return nil, errors.New("未识别到文本")
		}
		return nil, errors.New(rsp.Wording)
	}
	if rsp.RetCode != 0 {
		return nil, errors.Errorf("server error, code: %v msg: %v", rsp.RetCode, rsp.ErrMsg)
	}
	texts := make([]*TextDetection, 0, len(rsp.OcrRspBody.TextDetections))
	for _, text := range rsp.OcrRspBody.TextDetections {
		points := make([]*Coordinate, 0, len(text.Polygon.Coordinates))
		for _, c := range text.Polygon.Coordinates {
			points = append(points, &Coordinate{
				X: c.X,
				Y: c.Y,
			})
		}
		texts = append(texts, &TextDetection{
			Text:        text.DetectedText,
			Confidence:  text.Confidence,
			Coordinates: points,
		})
	}
	return &OcrResponse{
		Texts:    texts,
		Language: rsp.OcrRspBody.Language,
	}, nil
}
