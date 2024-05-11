package wtlogin

import (
	"errors"
	"strings"
	"unicode"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/system"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

// BuildRegisterRequest trpc.qq_new_tech.status_svc.StatusService.Register
func BuildRegisterRequest(app *auth.AppInfo, device *auth.DeviceInfo) []byte {
	return proto.DynamicMessage{
		1: strings.ToUpper(device.Guid),
		2: 0,
		3: app.CurrentVersion,
		4: 0,
		5: 2052,
		6: proto.DynamicMessage{
			1: device.DeviceName,
			2: capitalize(app.VendorOS),
			3: device.SystemKernel,
			4: "",
			5: app.VendorOS,
		},
		7: false, // set_mute
		8: false, // register_vendor_type
		9: true,  // regtype
	}.Encode()
}

// BuildSSOHeartbeatRequest trpc.qq_new_tech.status_svc.StatusService.SsoHeartBeat
func BuildSSOHeartbeatRequest() []byte {
	//return proto.DynamicMessage{
	//	1: 1,
	//}.Encode()
	// 直接硬编码罢
	return []byte{0x08, 0x01}
}

func ParseRegisterResponse(response []byte) error {
	var resp system.ServiceRegisterResponse
	err := proto.Unmarshal(response, &resp)
	if err != nil {
		return err
	}
	msg := resp.Message.Unwrap()
	if msg == "register success" {
		return nil
	}
	return errors.New(msg)
}

func capitalize(s string) string {
	news := make([]byte, len(s))
	rs := []rune(s)
	n := copy(news, string(unicode.ToUpper(rs[0])))
	copy(news[n:], strings.ToLower(s[n:]))
	return utils.B2S(news)
}
