package tlv

import (
	"fmt"
	"testing"

	"github.com/Redmomn/LagrangeGo/info"
)

func TestTlv(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	appinfo := info.AppList["linux"]
	deviceinfo := info.DeviceInfo{
		Guid:          "cfcd208495d565ef66e7dff9f98764da",
		DeviceName:    "Lagrange-DCFCD07E",
		SystemKernel:  "Windows 10.0.22631",
		KernelVersion: "10.0.22631",
	}
	// common
	t18 := T18(123, 123, 123, 123, 123, 123)
	t100 := T100(123, 123, 123, 123, 123, 123)
	// t106是正确的
	t106 := T106(123, 123, 123, "123456", data, data, data, true)
	t107 := T107(123, 123, 123, 123)
	t116 := T116(123)
	t124 := T124()
	t128 := T128("123", data)
	t141 := T141(data, data)
	t142 := T142("12341234123412341234123412341234", 123)
	t144 := T144(append(append(data, data...), append(data, data...)...), appinfo, deviceinfo)
	t145 := T145(data)
	t147 := T147(123, "123", "123")
	t166 := T166(123)
	t16a := T16a(data)
	t16e := T16e("123")
	t177 := T177("123", 123)
	t191 := T191(123)
	t318 := T318(data)
	t521 := T521(123, "123")
	fmt.Printf("t18: %x\n", t18)
	fmt.Printf("t100: %x\n", t100)
	fmt.Printf("t106: %x\n", t106)
	fmt.Printf("t107: %x\n", t107)
	fmt.Printf("t116: %x\n", t116)
	fmt.Printf("t124: %x\n", t124)
	fmt.Printf("t128: %x\n", t128)
	fmt.Printf("t141: %x\n", t141)
	fmt.Printf("t142: %x\n", t142)
	fmt.Printf("t144: %x\n", t144)
	fmt.Printf("t145: %x\n", t145)
	fmt.Printf("t147: %x\n", t147)
	fmt.Printf("t166: %x\n", t166)
	fmt.Printf("t16a: %x\n", t16a)
	fmt.Printf("t16e: %x\n", t16e)
	fmt.Printf("t177: %x\n", t177)
	fmt.Printf("t191: %x\n", t191)
	fmt.Printf("t318: %x\n", t318)
	fmt.Printf("t521: %x\n", t521)

	// qrcode
	t11 := T11(data)
	t16 := T16(123, 123, data, "123", "123")
	t1b := T1b()
	t1d := T1d(123)
	t33 := T33(data)
	t35 := T35(123)
	t66 := T66(123)
	td1 := Td1("123", "123")
	fmt.Printf("t11: %x\n", t11)
	fmt.Printf("t16: %x\n", t16)
	fmt.Printf("t1b: %x\n", t1b)
	fmt.Printf("t1d: %x\n", t1d)
	fmt.Printf("t33: %x\n", t33)
	fmt.Printf("t35: %x\n", t35)
	fmt.Printf("t66: %x\n", t66)
	fmt.Printf("td1: %x\n", td1)

}
