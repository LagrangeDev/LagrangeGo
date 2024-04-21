package highway

import (
	"encoding/hex"

	"github.com/LagrangeDev/LagrangeGo/utils/proto"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/action"
)

func BuildHighWayUrlReq(tgt []byte) ([]byte, error) {
	tgtHex := hex.EncodeToString(tgt)
	body := &action.HttpConn0X6Ff_501{
		HttpConn: &action.HttpConn{
			Field1:       0,
			Field2:       0,
			Field3:       16,
			Field4:       1,
			Tgt:          tgtHex,
			Field6:       3,
			ServiceTypes: []int32{1, 5, 10, 21},
			Field9:       2,
			Field10:      9,
			Field11:      8,
			Ver:          "1.0.1",
		},
	}
	packet, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func ParseHighWayUrlReq(data []byte) (*action.HttpConn0X6Ff_501Response, error) {
	var req action.HttpConn0X6Ff_501Response
	err := proto.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}
