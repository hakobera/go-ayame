package main

// #cgo LDFLAGS: -lvpx
// #include <stdlib.h>
// #include <stdint.h>
// #include <vpx/vpx_decoder.h>
// #include <vpx/vpx_image.h>
// #include <vpx/vp8dx.h>
//
// vpx_codec_iface_t *ifaceVP8() {
//   return vpx_codec_vp8_dx();
// }
// vpx_codec_iface_t *ifaceVP9() {
//   return vpx_codec_vp9_dx();
// }
// vpx_codec_ctx_t *newCtx() {
//   return malloc(sizeof(vpx_codec_ctx_t));
// }
//
// void yuv420_to_rgb(uint32_t width, uint32_t height,
// 					const uint8_t *y, const uint8_t *u, const uint8_t *v,
// 					int ystride, int ustride, int vstride,
// 					uint8_t *out)
// {
// 		unsigned long int i, j;
// 		for (i = 0; i < height; ++i) {
//	 		for (j = 0; j < width; ++j) {
//		 		uint8_t *point = out + 4 * ((i * width) + j);
//		 		int t_y = y[((i * ystride) + j)];
//		 		int t_u = u[(((i / 2) * ustride) + (j / 2))];
//		 		int t_v = v[(((i / 2) * vstride) + (j / 2))];
//		 		t_y = t_y < 16 ? 16 : t_y;
//
//		 		int r = (298 * (t_y - 16) + 409 * (t_v - 128) + 128) >> 8;
//		 		int g = (298 * (t_y - 16) - 100 * (t_u - 128) - 208 * (t_v - 128) + 128) >> 8;
//		 		int b = (298 * (t_y - 16) + 516 * (t_u - 128) + 128) >> 8;
//
//		 		point[0] = r>255? 255 : r<0 ? 0 : r;
//		 		point[1] = g>255? 255 : g<0 ? 0 : g;
//		 		point[2] = b>255? 255 : b<0 ? 0 : b;
//		 		point[3] = ~0;
//		 	}
//		}
// }
import "C"

import (
	"errors"
	"fmt"
	"image"
	"log"
	"reflect"
	"sync"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
)

func image2RGBA(img *C.vpx_image_t) *image.RGBA {
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
	)

	return &image.RGBA{
		Pix:    out,
		Stride: int(img.d_w) * 4,
		Rect:   image.Rect(0, 0, int(img.d_w), int(img.d_h)),
	}
}

type VpxFrame struct {
	Image      *image.RGBA
	IsKeyframe bool
}

type VpxDecoder struct {
	codec  *C.vpx_codec_ctx_t
	iface  *C.vpx_codec_iface_t
	format string

	mu     sync.Mutex
	closed bool
}

func NewDecoder(format string) (*VpxDecoder, error) {
	iface := (*C.vpx_codec_iface_t)(nil)
	switch format {
	case "VP8":
		iface = C.ifaceVP8()
	case "VP9":
		iface = C.ifaceVP9()
	default:
		return nil, fmt.Errorf("Invalid format: %s", format)
	}

	ctx := C.newCtx()
	d := &VpxDecoder{
		codec:  ctx,
		iface:  iface,
		format: format,

		closed: false,
	}
	log.Printf("Using %s\n", C.GoString(C.vpx_codec_iface_name(iface)))

	err := d.init()
	if err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func (d *VpxDecoder) NewSampleBuilder() *samplebuilder.SampleBuilder {
	var depacketizer rtp.Depacketizer = nil
	switch d.format {
	case "VP8":
		depacketizer = &codecs.VP8Packet{}
	case "VP9":
		depacketizer = &codecs.VP9Packet{}
	}
	return samplebuilder.New(10, depacketizer)
}

func (d *VpxDecoder) Process(src <-chan *media.Sample, out chan<- VpxFrame) {
	if d.closed {
		return
	}

	defer close(out)
	receiveFirstKeyFrame := false

	for pkt := range src {
		isKeyframe := (pkt.Data[0]&0x1 == 0)
		if !isKeyframe && !receiveFirstKeyFrame {
			continue
		}
		if isKeyframe && !receiveFirstKeyFrame {
			receiveFirstKeyFrame = true
		}

		err := d.decode(pkt)
		if err != nil {
			log.Println("[WARN]", err)
			continue
		}

		var iter C.vpx_codec_iter_t
		img := C.vpx_codec_get_frame(d.codec, &iter)
		for img != nil {
			out <- VpxFrame{
				Image:      image2RGBA(img),
				IsKeyframe: isKeyframe,
			}
			img = C.vpx_codec_get_frame(d.codec, &iter)
		}
	}
}

func (d *VpxDecoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}
	d.closed = true

	defer C.free(unsafe.Pointer(d.codec))

	if C.vpx_codec_destroy(d.codec) != 0 {
		return errors.New("vpx_codec_destroy failed")
	}
	return nil
}

func (d *VpxDecoder) init() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return errors.New("decoder is already closed")
	}

	ret := C.vpx_codec_dec_init_ver(d.codec, d.iface, nil, 0, C.VPX_DECODER_ABI_VERSION)
	if ret != 0 {
		return d.handleError("Failed to initialize decoder.")
	}
	return nil
}

func (d *VpxDecoder) decode(sample *media.Sample) error {
	p := (*reflect.SliceHeader)(unsafe.Pointer(&sample.Data))
	data := (*C.uchar)(unsafe.Pointer(p.Data))
	size := (C.uint)(len(sample.Data))
	ret := C.vpx_codec_decode(d.codec, data, size, nil, 0)
	if ret != 0 {
		return d.handleError("Failed to decode frame.")
	}
	return nil
}

func (d *VpxDecoder) handleError(msg string) error {
	code := C.GoString(C.vpx_codec_error(d.codec))
	detail := C.GoString(C.vpx_codec_error_detail(d.codec))
	return fmt.Errorf("%s %s : %s", msg, code, detail)
}
