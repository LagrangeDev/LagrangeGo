package wtlogin

import (
	"encoding/hex"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/login"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto/ecdh"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

var encKey, _ = hex.DecodeString("e2733bf403149913cbf80c7a95168bd4ca6935ee53cd39764beebe2e007e3aee")
var keyExangeLogger = utils.GetLogger("KeyExchange")

func BuildKexExchangeRequest(uin uint32, guid string) []byte {
	p1 := proto.DynamicMessage{
		1: uin,
		2: guid,
	}.Encode()

	encl := crypto.AesGCMEncrypt(p1, ecdh.Instance["prime256v1"].SharedKey())

	p2 := binary.NewBuilder(nil).
		WriteBytes(ecdh.Instance["prime256v1"].PublicKey(), false).
		WriteU32(1).
		WriteBytes(encl, false).
		WriteU32(0).
		WriteU32(uint32(utils.TimeStamp())).
		Pack(binary.PackTypeNone)

	p2Hash := utils.SHA256Digest(p2)
	encP2Hash := crypto.AesGCMEncrypt(p2Hash, encKey)

	return proto.DynamicMessage{
		1: ecdh.Instance["prime256v1"].PublicKey(),
		2: 1,
		3: encl,
		4: utils.TimeStamp(),
		5: encP2Hash,
	}.Encode()
}

func ParseKeyExchangeResponse(response []byte, sig *info.SigInfo) error {
	keyExangeLogger.Debugf("keyexchange proto data: %x", response)

	var p login.SsoKeyExchangeResponse
	err := proto.Unmarshal(response, &p)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}

	shareKey, err := ecdh.Instance["prime256v1"].Exange(p.PublicKey)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}

	var decPb login.SsoKeyExchangeDecrypted
	err = proto.Unmarshal(crypto.AesGCMDecrypt(p.GcmEncrypted, shareKey), &decPb)
	if err != nil {
		keyExangeLogger.Errorln(err)
		return err
	}

	sig.ExchangeKey = decPb.GcmKey
	sig.KeySig = decPb.Sign

	return nil
}
