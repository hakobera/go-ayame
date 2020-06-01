package vpx

// #include <stdlib.h>
// #include "binding.h"
import "C"

import (
	"image"
	"reflect"
	"unsafe"

	"github.com/hakobera/go-ayame/pkg/decoder"
)

type DecodedImage struct {
	isKeyFrame bool
	image      *C.vpx_image_t
}

func (f *DecodedImage) IsKeyFrame() bool {
	return f.isKeyFrame
}

func (f *DecodedImage) Width() uint32 {
	return uint32(f.image.d_w)
}

func (f *DecodedImage) Height() uint32 {
	return uint32(f.image.d_h)
}

func (f *DecodedImage) Plane(n int) []byte {
	var p *C.uchar

	switch n {
	case 0:
		p = f.image.planes[C.VPX_PLANE_Y]
	case 1:
		p = f.image.planes[C.VPX_PLANE_U]
	case 2:
		p = f.image.planes[C.VPX_PLANE_V]
	}

	lenData := f.Width() * f.Height()
	return (*[1 << 30]byte)(unsafe.Pointer(p))[:lenData:lenData]
}

func (f *DecodedImage) Stride(n int) int {
	switch n {
	case 0:
		return int(f.image.stride[C.VPX_PLANE_Y])
	case 1:
		return int(f.image.stride[C.VPX_PLANE_U])
	case 2:
		return int(f.image.stride[C.VPX_PLANE_V])
	default:
		return -1
	}
}

func (f *DecodedImage) ToBytes(format decoder.ColorFormat) []uint8 {
	img := f.image
	out := make([]uint8, img.d_w*img.d_h*4)

	C.yuv420_to_rgb(
		img.d_w,
		img.d_h,
		img.planes[C.VPX_PLANE_Y],
		img.planes[C.VPX_PLANE_U],
		img.planes[C.VPX_PLANE_V],
		img.stride[C.VPX_PLANE_Y],
		img.stride[C.VPX_PLANE_U],
		img.stride[C.VPX_PLANE_V],
		(*C.uint8_t)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&out)).Data)),
		(C.int)(format),
	)

	return out
}

func (f *DecodedImage) ToRGBA() *image.RGBA {
	out := f.ToBytes(decoder.ColorRGBA)

	return &image.RGBA{
		Pix:    out,
		Stride: int(f.image.d_w) * 4,
		Rect:   image.Rect(0, 0, int(f.image.d_w), int(f.image.d_h)),
	}
}
