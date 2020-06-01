package vpx

// #include <stdlib.h>
// #include "binding.h"
import "C"

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"unsafe"

	"github.com/hakobera/go-ayame/pkg/decoder"
	"github.com/pion/rtp/codecs"
)

type VP8Decoder struct {
	context *C.vpx_codec_ctx_t

	lastFrameWidth  int
	lastFrameHeight int
	initialized     bool
	closed          bool
	mu              sync.Mutex
}

func NewVP8Decoder() (*VP8Decoder, error) {
	ctx := C.newVpxCtx()
	d := &VP8Decoder{
		context:     ctx,
		initialized: false,
		closed:      false,
	}

	err := d.init()
	if err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func (d *VP8Decoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}
	d.closed = true

	defer C.free(unsafe.Pointer(d.context))

	if C.vpx_codec_destroy(d.context) != 0 {
		return errors.New("vpx_codec_destroy failed")
	}
	d.initialized = false
	return nil
}

func (d *VP8Decoder) NewFrameBuilder() *decoder.FrameBuilder {
	return decoder.NewFrameBuilder(10, &codecs.VP8Packet{})
}

func (d *VP8Decoder) Process(src <-chan *decoder.Frame, out chan<- decoder.DecodedImage) {
	if d.closed {
		return
	}

	defer close(out)

	keyFrameRequied := true

	for pkt := range src {
		var err error
		var f []byte

		f, err = d.assembleFrame(pkt.Data)
		if err != nil {
			log.Println("[WARN]", err)
			continue
		}

		isKeyFrame := isVP8KeyFrame(f)
		if keyFrameRequied {
			if !isKeyFrame {
				continue
			}
			keyFrameRequied = false
		}

		err = d.decode(f)
		if err != nil {
			log.Println("[WARN]", err)
			continue
		}

		var img *C.vpx_image_t
		var iter C.vpx_codec_iter_t

		img = C.vpx_codec_get_frame(d.context, &iter)
		for img != nil {
			out <- &DecodedImage{
				image:      img,
				isKeyFrame: isKeyFrame,
			}
			img = C.vpx_codec_get_frame(d.context, &iter)
		}
	}
}

func (d *VP8Decoder) init() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return errors.New("docoder is already closed")
	}

	config := C.newVpxDecCfg()
	defer C.free(unsafe.Pointer(config))
	config.threads = 1
	config.h = 0
	config.w = 0

	ret := C.vpx_codec_dec_init_ver(d.context, C.ifaceVP8(), config, C.VPX_CODEC_USE_POSTPROC, C.VPX_DECODER_ABI_VERSION)
	if ret != C.VPX_CODEC_OK {
		return handleError(d.context, "Feiled to initialize decoder.")
	}
	d.initialized = true
	return nil
}

func (d *VP8Decoder) decode(frame []byte) error {
	ppcfg := C.newVP8PostProcCfg()
	defer C.free(unsafe.Pointer(ppcfg))

	ppcfg.post_proc_flag = C.VP8_MFQE | C.VP8_DEBLOCK
	if d.lastFrameWidth*d.lastFrameHeight <= 640*360 {
		ppcfg.post_proc_flag |= C.VP8_DEMACROBLOCK
	}
	ppcfg.deblocking_level = 3
	C.vpxCodecControl(d.context, C.VP8_SET_POSTPROC, unsafe.Pointer(ppcfg))

	p := (*reflect.SliceHeader)(unsafe.Pointer(&frame))
	data := (*C.uchar)(unsafe.Pointer(p.Data))
	size := (C.uint)(len(frame))
	ret := C.vpx_codec_decode(d.context, data, size, nil, 0)
	if ret != C.VPX_CODEC_OK {
		return handleError(d.context, "Failed to decode frame.")
	}
	return nil
}

func (d *VP8Decoder) assembleFrame(data [][]byte) ([]byte, error) {
	var a []byte
	for _, d := range data {
		a = append(a, d...)
	}
	return a, nil
}

func isVP8KeyFrame(frame []byte) bool {
	return (frame[0]&0x1 == 0)
}
