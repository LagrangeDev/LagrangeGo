// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/service/oidb/OidbSvcTrpcTcp0xE07_0.proto

package oidb

type OidbSvcTrpcTcp0XE07_0 struct {
	Version    uint32      `protobuf:"varint,1,opt"`
	Client     uint32      `protobuf:"varint,2,opt"`
	Entrance   uint32      `protobuf:"varint,3,opt"`
	OcrReqBody *OcrReqBody `protobuf:"bytes,10,opt"`
	_          [0]func()
}

type OcrReqBody struct {
	ImageUrl              string `protobuf:"bytes,1,opt"`
	LanguageType          uint32 `protobuf:"varint,2,opt"`
	Scene                 uint32 `protobuf:"varint,3,opt"`
	OriginMd5             string `protobuf:"bytes,10,opt"`
	AfterCompressMd5      string `protobuf:"bytes,11,opt"`
	AfterCompressFileSize string `protobuf:"bytes,12,opt"`
	AfterCompressWeight   string `protobuf:"bytes,13,opt"`
	AfterCompressHeight   string `protobuf:"bytes,14,opt"`
	IsCut                 bool   `protobuf:"varint,15,opt"`
	_                     [0]func()
}

type OidbSvcTrpcTcp0XE07_0_Response struct {
	RetCode    int32       `protobuf:"varint,1,opt"`
	ErrMsg     string      `protobuf:"bytes,2,opt"`
	Wording    string      `protobuf:"bytes,3,opt"`
	OcrRspBody *OcrRspBody `protobuf:"bytes,10,opt"`
	_          [0]func()
}

type OcrRspBody struct {
	TextDetections           []*TextDetection `protobuf:"bytes,1,rep"`
	Language                 string           `protobuf:"bytes,2,opt"`
	RequestId                string           `protobuf:"bytes,3,opt"`
	OcrLanguageList          []string         `protobuf:"bytes,101,rep"`
	DstTranslateLanguageList []string         `protobuf:"bytes,102,rep"`
	LanguageList             []*Language      `protobuf:"bytes,103,rep"`
	AfterCompressWeight      uint32           `protobuf:"varint,111,opt"`
	AfterCompressHeight      uint32           `protobuf:"varint,112,opt"`
}

type TextDetection struct {
	DetectedText string   `protobuf:"bytes,1,opt"`
	Confidence   int32    `protobuf:"varint,2,opt"`
	Polygon      *Polygon `protobuf:"bytes,3,opt"`
	AdvancedInfo string   `protobuf:"bytes,4,opt"`
	_            [0]func()
}

type Polygon struct {
	Coordinates []*Coordinate `protobuf:"bytes,1,rep"`
}

type Coordinate struct {
	X int32 `protobuf:"varint,1,opt"`
	Y int32 `protobuf:"varint,2,opt"`
	_ [0]func()
}

type Language struct {
	LanguageCode string `protobuf:"bytes,1,opt"`
	LanguageDesc string `protobuf:"bytes,2,opt"`
	_            [0]func()
}
