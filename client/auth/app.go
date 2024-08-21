package auth

import (
	"encoding/hex"

	"github.com/LagrangeDev/LagrangeGo/internal/proto"
)

var AppList = map[string]map[string]*AppInfo{
	"linux": {
		"3.1.2-13107": {
			OS:       "Linux",
			Kernel:   "Linux",
			VendorOS: "linux",

			CurrentVersion:   "3.1.2-13107",
			BuildVersion:     13107,
			MiscBitmap:       32764,
			PTVersion:        "2.0.0",
			PTOSVersion:      19,
			PackageName:      "com.tencent.qq",
			WTLoginSDK:       "nt.wtlogin.0.0.1",
			PackageSign:      "V1_LNX_NQ_3.1.2-13107_RDM_B",
			AppID:            1600001615,
			SubAppID:         537146866,
			AppIDQrcode:      13697054,
			AppClientVersion: 13172,

			MainSigmap:  169742560,
			SubSigmap:   0,
			NTLoginType: 1,
		},

		"3.2.10-25765": {
			OS:       "Linux",
			Kernel:   "Linux",
			VendorOS: "linux",

			CurrentVersion:   "3.2.10-25765",
			BuildVersion:     25765,
			MiscBitmap:       32764,
			PTVersion:        "2.0.0",
			PTOSVersion:      19,
			PackageName:      "com.tencent.qq",
			WTLoginSDK:       "nt.wtlogin.0.0.1",
			PackageSign:      "V1_LNX_NQ_3.2.10_25765_GW_B",
			AppID:            1600001615,
			SubAppID:         537234773,
			AppIDQrcode:      13697054,
			AppClientVersion: 13172,

			MainSigmap:  169742560,
			SubSigmap:   0,
			NTLoginType: 1,
		},
	},

	"macos": {
		"6.9.20-17153": {
			OS:       "Mac",
			Kernel:   "Darwin",
			VendorOS: "mac",

			CurrentVersion:   "6.9.20-17153",
			BuildVersion:     17153,
			MiscBitmap:       32764,
			PTVersion:        "2.0.0",
			PTOSVersion:      23,
			PackageName:      "com.tencent.qq",
			WTLoginSDK:       "nt.wtlogin.0.0.1",
			PackageSign:      "V1_MAC_NQ_6.9.20-17153_RDM_B",
			AppID:            1600001602,
			SubAppID:         537162356,
			AppIDQrcode:      537162356,
			AppClientVersion: 13172,

			MainSigmap:  169742560,
			SubSigmap:   0,
			NTLoginType: 5,
		},
	},
	"windows": {
		"9.9.12-25493": {
			OS:       "Windows",
			Kernel:   "Windows_NT",
			VendorOS: "win32",

			CurrentVersion:   "9.9.12-25493",
			BuildVersion:     25493,
			MiscBitmap:       32764,
			PTVersion:        "2.0.0",
			PTOSVersion:      23,
			PackageName:      "com.tencent.qq",
			WTLoginSDK:       "nt.wtlogin.0.0.1",
			PackageSign:      "V1_WIN_NQ_9.9.12-25493_GW_B",
			AppID:            1600001604,
			SubAppID:         537231759,
			AppIDQrcode:      537138217,
			AppClientVersion: 13172,

			MainSigmap:  169742560,
			SubSigmap:   0,
			NTLoginType: 5,
		},
	},
}

type AppInfo struct {
	OS       string `json:"os"`
	Kernel   string `json:"kernel"`
	VendorOS string `json:"vendor_os"`

	CurrentVersion   string `json:"current_version"`
	BuildVersion     int    `json:"build_version"`
	MiscBitmap       int    `json:"misc_bitmap"`
	PTVersion        string `json:"pt_version"`
	PTOSVersion      int    `json:"pt_os_version"`
	PackageName      string `json:"package_name"`
	WTLoginSDK       string `json:"wtlogin_sdk"`
	PackageSign      string `json:"package_sign"`
	AppID            int    `json:"app_id"`
	SubAppID         int    `json:"sub_app_id"`
	AppIDQrcode      int    `json:"app_id_qrcode"`
	AppClientVersion int    `json:"app_client_version"`
	MainSigmap       int    `json:"main_sigmap"`
	SubSigmap        int    `json:"sub_sigmap"`
	NTLoginType      int    `json:"nt_login_type"`

	SignExtraHex string `json:"sign_extra_hex"`
}

func init() {
	for _, appOs := range AppList {
		for _, app := range appOs {
			app.SignExtraHex = hex.EncodeToString(proto.DynamicMessage{2: app.PackageSign}.Encode())
		}
	}
}
