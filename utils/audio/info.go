package audio

// https://github.com/LagrangeDev/lagrange-python/tree/broken/lagrange/utils/audio

import (
	binary2 "encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	lgrio "github.com/LagrangeDev/LagrangeGo/utils/io"
)

type Info struct {
	Type Type
	Time float32
}

func decode(r io.ReadSeeker, _f bool) (*Info, error) {
	reader := binary.ParseReader(r)
	buf := reader.ReadBytes(1)
	if len(buf) != 1 || buf[0] != 0x23 {
		if !_f {
			return decode(r, true)
		}
		return nil, errors.New("unknown audio type")
	}
	buf = append(buf, reader.ReadBytes(5)...)

	switch {
	case strings.HasPrefix(string(buf), "#!AMR\n"):
		return &Info{
			Type: amr,
			Time: float32(len(reader.ReadAll())) / 1607.0,
		}, nil
	case string(buf) == "#!SILK":
		ver := reader.ReadBytes(3)
		if string(ver) != "_V3" {
			return nil, fmt.Errorf("unsupported silk version: %s", lgrio.B2S(ver))
		}
		data := reader.ReadAll()
		size := len(data)

		var typ Type
		if _f { // txsilk
			typ = txSilk
		} else {
			typ = silkV3
		}

		blks := 0
		pos := 0

		for pos+2 < size {
			length := binary2.LittleEndian.Uint16(data[pos : pos+2])
			if length == 0xFFFF {
				break
			}
			blks++
			pos += int(length) + 2
		}
		return &Info{
			Type: typ,
			Time: float32(blks) * 0.02,
		}, nil
	default:
		return nil, errors.New("unknown audio type")
	}
}

func Decode(r io.ReadSeeker) (*Info, error) {
	defer func() {
		_, _ = r.Seek(0, io.SeekStart)
	}()
	_, _ = r.Seek(0, io.SeekStart)
	return decode(r, false)
}
