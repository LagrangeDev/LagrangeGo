package wtlogin

import (
	"strings"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/system"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

// BuildRegisterRequest trpc.qq_new_tech.status_svc.StatusService.Register
func BuildRegisterRequest(app *info.AppInfo, device *info.DeviceInfo) []byte {
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

func ParseRegisterResponse(response []byte) bool {
	var resp system.ServiceRegisterResponse
	err := proto.Unmarshal(response, &resp)
	if err != nil {
		loginLogger.Errorf("Unmarshal register response failed: %s", err)
		return false
	}
	if resp.Message.Unwrap() == "register success" {
		return true
	}
	return false
}

func capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}
