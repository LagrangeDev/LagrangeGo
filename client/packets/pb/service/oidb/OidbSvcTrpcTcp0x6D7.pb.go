// Code generated by protoc-gen-golite. DO NOT EDIT.
// source: pb/service/oidb/OidbSvcTrpcTcp0x6D7.proto

package oidb

type OidbSvcTrpcTcp0X6D7 struct {
	Delete *OidbSvcTrpcTcp0X6D7Delete `protobuf:"bytes,2,opt"`
	_      [0]func()
}

type OidbSvcTrpcTcp0X6D7Delete struct {
	GroupUin uint32 `protobuf:"varint,1,opt"`
	FolderId string `protobuf:"bytes,3,opt"`
	_        [0]func()
}

type OidbSvcTrpcTcp0X6D7Response struct {
	Delete *OidbSvcTrpcTcp0X6D7_1Response `protobuf:"bytes,2,opt"`
	_      [0]func()
}

type OidbSvcTrpcTcp0X6D7_1Response struct {
	RetCode       int32  `protobuf:"varint,1,opt"`
	RetMsg        string `protobuf:"bytes,2,opt"`
	ClientWording string `protobuf:"bytes,3,opt"`
	_             [0]func()
}