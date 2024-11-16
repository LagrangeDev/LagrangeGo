package utils

import (
	"errors"
	"io"

	"github.com/fumiama/imgsz"
)

var (
	ErrImageDataTooShort = errors.New("image data is too short")
)

type ImageFormat uint32

const (
	Unknown ImageFormat = 0000
	Jpeg    ImageFormat = 1000
	Png     ImageFormat = 1001
	Gif     ImageFormat = 2000
	Webp    ImageFormat = 1002
	Bmp     ImageFormat = 1005
	Tiff    ImageFormat = 1006
)

var formatmap = map[string]ImageFormat{
	"jpeg": Jpeg,
	"png":  Png,
	"gif":  Gif,
	"webp": Webp,
	"bmp":  Bmp,
	"tiff": Tiff,
}

func (format ImageFormat) String() string {
	//nolint
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

func ImageResolve(image io.ReadSeeker) (format ImageFormat, size imgsz.Size, err error) {
	defer func(image io.ReadSeeker, offset int64, whence int) {
		_, _ = image.Seek(offset, whence)
	}(image, 0, io.SeekStart)
	if _, err = image.Seek(10, io.SeekStart); err != nil { // 最小长度检查
		err = ErrImageDataTooShort
		return
	}
	_, _ = image.Seek(0, io.SeekStart)
	sz, fmts, err := imgsz.DecodeSize(image)
	if err != nil {
		return
	}

	return formatmap[fmts], sz, nil
}
