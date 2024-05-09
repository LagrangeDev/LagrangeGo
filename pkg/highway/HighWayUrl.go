package highway

import (
	"encoding/hex"

	"github.com/LagrangeDev/LagrangeGo/utils/proto"

	"github.com/LagrangeDev/LagrangeGo/pkg/pb/action"
)

func BuildHighWayUrlReq(tgt []byte) ([]byte, error) {
	return proto.Marshal(&action.HttpConn0X6Ff_501{
		HttpConn: &action.HttpConn{
			Field1:       0,
			Field2:       0,
			Field3:       16,
			Field4:       1,
			Tgt:          hex.EncodeToString(tgt),
			Field6:       3,
			ServiceTypes: []int32{1, 5, 10, 21},
			Field9:       2,
			Field10:      9,
			Field11:      8,
			Ver:          "1.0.1",
		},
	})
}

func ParseHighWayUrlReq(data []byte) (req action.HttpConn0X6Ff_501Response, err error) {
	err = proto.Unmarshal(data, &req)
	return
}
