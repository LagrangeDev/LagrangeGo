package wtlogin

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/login"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto/ecdh"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

var encKey, _ = hex.DecodeString("e2733bf403149913cbf80c7a95168bd4ca6935ee53cd39764beebe2e007e3aee")

func BuildKexExchangeRequest(uin uint32, guid string) ([]byte, error) {
	plain1, err := proto.Marshal(&login.SsoKeyExchangePlain{
		Uin:  proto.Some(strconv.Itoa(int(uin))),
		Guid: io.MustParseHexStr(guid),
	})
	if err != nil {
		return nil, err
	}

	encl, err := crypto.AESGCMEncrypt(plain1, ecdh.P256().SharedKey())
	if err != nil {
		return nil, err
	}

	p2Hash := crypto.SHA256Digest(
		binary.NewBuilder().
			WriteBytes(ecdh.P256().PublicKey()).
			WriteU32(1).
			WriteBytes(encl).
			WriteU32(0).
			WriteU32(uint32(io.TimeStamp())).
			ToBytes(),
	)
	encP2Hash, err := crypto.AESGCMEncrypt(p2Hash, encKey)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&login.SsoKeyExchange{
		PubKey:    ecdh.P256().PublicKey(),
		Type:      1,
		GcmCalc1:  encl,
		Timestamp: uint32(time.Now().Unix()),
		GcmCalc2:  encP2Hash,
	})
}

func ParseKeyExchangeResponse(response []byte, sig *auth.SigInfo) error {
	var p login.SsoKeyExchangeResponse
	err := proto.Unmarshal(response, &p)
	if err != nil {
		return err
	}

	shareKey, err := ecdh.P256().Exange(p.PublicKey)
	if err != nil {
		return err
	}

	var decPb login.SsoKeyExchangeDecrypted
	data, err := crypto.AESGCMDecrypt(p.GcmEncrypted, shareKey)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(data, &decPb)
	if err != nil {
		return err
	}
	sig.ExchangeKey = decPb.GcmKey
	sig.KeySig = decPb.Sign

	return nil
}
