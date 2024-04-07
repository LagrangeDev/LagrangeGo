package wtlogin

import (
	"encoding/hex"

	"github.com/Redmomn/LagrangeGo/packets/pb/login"

	"github.com/Redmomn/LagrangeGo/info"
	"github.com/Redmomn/LagrangeGo/utils"
	"github.com/Redmomn/LagrangeGo/utils/binary"
	"github.com/Redmomn/LagrangeGo/utils/crypto"
	"github.com/Redmomn/LagrangeGo/utils/crypto/ecdh"
	"github.com/Redmomn/LagrangeGo/utils/proto"
)

var encKey, _ = hex.DecodeString("e2733bf403149913cbf80c7a95168bd4ca6935ee53cd39764beebe2e007e3aee")
var keyExangeLogger = utils.GetLogger("KeyExchange")

func BuildKexExchangeRequest(uin uint32, guid string) []byte {
	p1 := proto.DynamicMessage{
		1: uin,
		2: guid,
	}.Encode()

	encl := crypto.AesGCMEncrypt(p1, ecdh.ECDH["prime256v1"].GetShareKey())

	p2 := binary.NewBuilder(nil).
		WriteBytes(ecdh.ECDH["prime256v1"].GetPublicKey(), false).
		WriteU32(1).
		WriteBytes(encl, false).
		WriteU32(0).
		WriteU32(uint32(utils.TimeStamp())).
		Pack(-1)

	p2Hash := utils.Sha256Digest(p2)
	encP2Hash := crypto.AesGCMEncrypt(p2Hash, encKey)

	return proto.DynamicMessage{
		1: ecdh.ECDH["prime256v1"].GetPublicKey(),
		2: 1,
		3: encl,
		4: utils.TimeStamp(),
		5: encP2Hash,
	}.Encode()
}

func ParseKeyExchangeResponse(response []byte, sig *info.SigInfo) {
	keyExangeLogger.Debugf("keyexchange proto data: %x", response)
	var p login.SsoKeyExchangeResponse
	err := proto.Unmarshal(response, p)
	if err != nil {
		keyExangeLogger.Error(err)
	}

	shareKey := ecdh.ECDH["prime256v1"].Exange(p.PublicKey)
	var decPb login.SsoKeyExchangeDecrypted
	err = proto.Unmarshal(crypto.AesGCMDecrypt(p.GcmEncrypted, shareKey), &decPb)
	if err != nil {
		keyExangeLogger.Error(err)
	}

	sig.ExchangeKey = decPb.GcmKey
	sig.KeySig = decPb.Sign
}
