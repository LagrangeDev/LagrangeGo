package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Vector2 二维向量
type Vector2 struct {
	X uint32
	Y uint32
}

type ImageFormat uint32

const (
	Unknown ImageFormat = iota
	Jpeg    ImageFormat = 1000
	Png     ImageFormat = 1001
	Gif     ImageFormat = 2000
	Webp    ImageFormat = 1002
	Bmp     ImageFormat = 1005
	Tiff    ImageFormat = 1006
)

func GetImageExt(format ImageFormat) string {
	switch format {
	case Jpeg:
		return "jpg"
	case Png:
		return "png"
	case Gif:
		return "gif"
	case Webp:
		return "webp"
	case Bmp:
		return "bmp"
	case Tiff:
		return "tiff"
	default:
		return "unknown"
	}
}

func ImageResolve(image []byte) (ImageFormat, Vector2, error) {
	if len(image) < 10 { // 最小长度检查
		return Unknown, Vector2{}, errors.New("image data is too short")
	}

	size := Vector2{}
	format := Unknown

	switch {
	case bytes.Equal(image[:6], []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}) || bytes.Equal(image[:6], []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}): // GIF
		size = Vector2{X: uint32(binary.LittleEndian.Uint16(image[6:8])), Y: uint32(binary.LittleEndian.Uint16(image[8:10]))}
		format = Gif

	case bytes.Equal(image[:2], []byte{0xFF, 0xD8}): // JPEG
		for i := 2; i < len(image)-10; i++ {
			if binary.LittleEndian.Uint16(image[i:i+2])&0xFCFF == 0xC0FF { // SOF0 ~ SOF3
				size = Vector2{X: uint32(binary.BigEndian.Uint16(image[i+7 : i+9])), Y: uint32(binary.BigEndian.Uint16(image[i+5 : i+7]))}
				break
			}
		}
		format = Jpeg

	case bytes.Equal(image[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}): // PNG
		size = Vector2{X: binary.BigEndian.Uint32(image[16:20]), Y: binary.BigEndian.Uint32(image[20:24])}
		format = Png

	case bytes.Equal(image[:4], []byte{0x52, 0x49, 0x46, 0x46}) && bytes.Equal(image[8:12], []byte{0x57, 0x45, 0x42, 0x50}): // WEBP
		if bytes.Equal(image[12:16], []byte{0x56, 0x50, 0x38, 0x58}) { // VP8X
			size = Vector2{X: uint32(binary.LittleEndian.Uint16(image[24:27]) + 1), Y: uint32(binary.LittleEndian.Uint16(image[27:30]) + 1)}
		} else if bytes.Equal(image[12:16], []byte{0x56, 0x50, 0x38, 0x4C}) { // VP8L
			size = Vector2{X: uint32(int32(binary.LittleEndian.Uint32(image[21:25]))&0x3FFF) + 1, Y: uint32(int32(binary.LittleEndian.Uint32(image[20:22])&0xFFFC000)>>0x0E) + 1}
		} else {
			size = Vector2{X: uint32(binary.LittleEndian.Uint16(image[26:28])), Y: uint32(binary.LittleEndian.Uint16(image[28:30]))}
		}
		format = Webp

	case bytes.Equal(image[:2], []byte{0x42, 0x4D}): // BMP
		size = Vector2{X: binary.LittleEndian.Uint32(image[18:22]), Y: binary.LittleEndian.Uint32(image[22:26])}
		format = Bmp

	case bytes.Equal(image[:2], []byte{0x49, 0x49}) || bytes.Equal(image[:2], []byte{0x4D, 0x4D}): // TIFF
		size = Vector2{X: uint32(binary.LittleEndian.Uint16(image[18:20])), Y: uint32(binary.LittleEndian.Uint16(image[30:32]))}
		format = Tiff

	default:
		return Unknown, Vector2{}, nil
	}

	return format, size, nil
}
