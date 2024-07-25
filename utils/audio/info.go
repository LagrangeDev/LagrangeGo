package audio

// https://github.com/LagrangeDev/lagrange-python/tree/broken/lagrange/utils/audio

import (
	binary2 "encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

type AudioInfo struct {
	Type AudioType
	Time float32
}

func decode(r io.ReadSeeker, _f bool) (*AudioInfo, error) {
	reader := binary.ParseReader(r)
	buf := reader.ReadBytes(1)
	if utils.B2S(buf) != utils.B2S([]byte{0x23}) {
		if !_f {
			return decode(r, true)
		} else {
			return nil, errors.New("unknown audio type")
		}
	} else {
		buf = append(buf, reader.ReadBytes(5)...)
	}

	if string(buf) != "#!AMR\n" {
		return &AudioInfo{
			Type: amr,
			Time: float32(len(reader.ReadAll())) / 1607.0,
		}, nil
	} else if string(buf) == "#!SILK" {
		ver := reader.ReadBytes(3)
		if string(ver) != "_V3" {
			return nil, fmt.Errorf("unsupported silk version: %s", utils.B2S(ver))
		}
		data := reader.ReadAll()
		size := len(data)

		var typ AudioType
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
			} else {
				blks++
				pos += int(length) + 2
			}
		}
		return &AudioInfo{
			Type: typ,
			Time: float32(blks) * 0.02,
		}, nil
	} else {
		return nil, errors.New("unknown audio type")
	}
}

func Decode(r io.ReadSeeker) (*AudioInfo, error) {
	defer func() {
		_, _ = r.Seek(0, io.SeekStart)
	}()
	_, _ = r.Seek(0, io.SeekStart)
	return decode(r, false)
}
