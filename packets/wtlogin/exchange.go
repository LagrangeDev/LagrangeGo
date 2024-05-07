package wtlogin

import (
	"encoding/hex"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/login"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto/ecdh"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

var encKey, _ = hex.DecodeString("e2733bf403149913cbf80c7a95168bd4ca6935ee53cd39764beebe2e007e3aee")
var keyExangeLogger = utils.GetLogger("KeyExchange")

func BuildKexExchangeRequest(uin uint32, guid string) ([]byte, error) {
	p1 := proto.DynamicMessage{
		1: uin,
		2: guid,
	}.Encode()

	encl, err := crypto.AesGCMEncrypt(p1, ecdh.P256().SharedKey())
	if err != nil {
		return nil, err
	}

	p2 := binary.NewBuilder(nil).
		WriteBytes(ecdh.P256().PublicKey(), false).
		WriteU32(1).
		WriteBytes(encl, false).
		WriteU32(0).
		WriteU32(uint32(utils.TimeStamp())).
		ToBytes()

	p2Hash := crypto.SHA256Digest(p2)
	encP2Hash, err := crypto.AesGCMEncrypt(p2Hash, encKey)
	if err != nil {
		return nil, err
	}

	return proto.DynamicMessage{
		1: ecdh.P256().PublicKey(),
		2: 1,
		3: encl,
		4: utils.TimeStamp(),
		5: encP2Hash,
	}.Encode(), nil
}

func ParseKeyExchangeResponse(response []byte, sig *info.SigInfo) error {
	keyExangeLogger.Debugf("keyexchange proto data: %x", response)

	var p login.SsoKeyExchangeResponse
	err := proto.Unmarshal(response, &p)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}

	shareKey, err := ecdh.P256().Exange(p.PublicKey)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}

	var decPb login.SsoKeyExchangeDecrypted
	data, err := crypto.AesGCMDecrypt(p.GcmEncrypted, shareKey)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}
	err = proto.Unmarshal(data, &decPb)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}

	sig.ExchangeKey = decPb.GcmKey
	sig.KeySig = decPb.Sign

	return nil
}
