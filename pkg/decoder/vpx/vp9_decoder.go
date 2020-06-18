package vpx

/*
#include <stdlib.h>
#include "binding.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/hakobera/go-ayame/pkg/decoder"
	gopointer "github.com/mattn/go-pointer"
	"github.com/pion/rtp/codecs"
)

func isVP9KeyFrame(packet []byte) bool {
	p := &codecs.VP9Packet{}
	if _, err := p.Unmarshal(packet); err != nil {
		return false
	}
	return !p.P
}

type VP9Decoder struct {
	context *C.vpx_codec_ctx_t

	frameBufferPool *VP9FrameBufferPool
	initialized     bool
	mu              sync.Mutex
}

func NewVP9Decoder() (*VP9Decoder, error) {
	ctx := C.newVpxCtx()
	d := &VP9Decoder{
		context:     ctx,
		initialized: false,
	}

	err := d.init()
	if err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func (d *VP9Decoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.initialized = false

	if d.frameBufferPool != nil {
		d.frameBufferPool.Clear()
	}

	if d.context != nil {
		if C.vpx_codec_destroy(d.context) != 0 {
			return errors.New("vpx_codec_destroy failed")
		}
		C.free(unsafe.Pointer(d.context))
		d.context = nil
	}
	return nil
}

func (d *VP9Decoder) NewFrameBuilder() *decoder.FrameBuilder {
	return decoder.NewFrameBuilder(10, &VP9Packet{}, &VP9PartitionHeadChecker{})
}

func (d *VP9Decoder) Process(src <-chan *decoder.Frame, out chan<- decoder.DecodedImage) {
	if !d.initialized {
		return
	}

	defer close(out)

	// TODO: Implement keyframe detection logic correctly
	keyFrameRequied := false

	for pkt := range src {
		var err error = nil
		var f []byte

		f, err = d.assembleFrame(pkt.Data)
		if err != nil {
			log.Println("[WARN]", err)
			continue
		}

		isKeyFrame := isVP9KeyFrame(pkt.FirstPacket)
		if keyFrameRequied {
			if !isKeyFrame {
				continue
			}
			keyFrameRequied = false
		}

		err = d.decode(f)
		if err != nil {
			//log.Println("[WARN]", err)
			continue
		}

		var img *C.vpx_image_t
		var iter C.vpx_codec_iter_t
		var qp C.int

		img = C.vpx_codec_get_frame(d.context, &iter)
		for img != nil {
			ret := C.vpxCodecControl(d.context, C.VPXD_GET_LAST_QUANTIZER, unsafe.Pointer(&qp))
			if ret != C.VPX_CODEC_OK {
				break
			}

			out <- &DecodedImage{
				image:      img,
				isKeyFrame: isKeyFrame,
			}
			img = C.vpx_codec_get_frame(d.context, &iter)
		}
	}
}

func (d *VP9Decoder) init() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	config := C.newVpxDecCfg()
	defer C.free(unsafe.Pointer(config))
	config.threads = (C.uint)(min(runtime.NumCPU(), 8))

	ret := C.vpx_codec_dec_init_ver(d.context, C.ifaceVP9(), config, 0, C.VPX_DECODER_ABI_VERSION)
	if ret != C.VPX_CODEC_OK {
		return handleError(d.context, "Feiled to initialize decoder.")
	}

	d.frameBufferPool = &VP9FrameBufferPool{}
	err := d.frameBufferPool.Init(d.context)
	if err != nil {
		return err
	}
	d.initialized = true
	return nil
}

func (d *VP9Decoder) decode(frame []byte) error {
	var data *C.uchar = nil

	p := (*reflect.SliceHeader)(unsafe.Pointer(&frame))
	size := (C.uint)(len(frame))
	if size > 0 {
		data = (*C.uchar)(unsafe.Pointer(p.Data))
	}
	ret := C.vpx_codec_decode(d.context, data, size, nil, 1 /*VPX_DL_REALTIME*/)
	if ret != C.VPX_CODEC_OK {
		return handleError(d.context, "Failed to decode frame.")
	}
	return nil
}

func (d *VP9Decoder) assembleFrame(data [][]byte) ([]byte, error) {
	var a []byte
	for _, d := range data {
		a = append(a, d...)
	}
	return a, nil
}

type VP9FrameBufferPool struct {
	this             unsafe.Pointer
	allocatedBuffers []*VP9FrameBuffer
	mu               sync.Mutex
}

func (p *VP9FrameBufferPool) Init(ctx *C.vpx_codec_ctx_t) error {
	p.this = gopointer.Save(p)
	ret := C.vpxCodecSetFrameBufferFunction(ctx, p.this)
	if ret != C.VPX_CODEC_OK {
		p.Clear()
		return fmt.Errorf("failed to initialize VP9FrameBufferPool")
	}
	p.allocatedBuffers = make([]*VP9FrameBuffer, 0)
	return nil
}

func (p *VP9FrameBufferPool) GetFrameBuffer(minSize C.size_t) (*VP9FrameBuffer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if minSize < 1 {
		return nil, fmt.Errorf("minSize must be greater than zero")
	}

	fmt.Println("GetFrameBuffer")

	var availableBuffer *VP9FrameBuffer = nil

	for _, buf := range p.allocatedBuffers {
		if buf.HasOneRef() {
			availableBuffer = buf
			break
		}
	}

	if availableBuffer == nil {
		availableBuffer = &VP9FrameBuffer{}
		availableBuffer.AddRef()
		p.allocatedBuffers = append(p.allocatedBuffers, availableBuffer)
	}

	availableBuffer.SetSize(minSize)
	return availableBuffer, nil
}

func (p *VP9FrameBufferPool) GetNumBuffersInUse() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := 0
	for _, buf := range p.allocatedBuffers {
		if !buf.HasOneRef() {
			n++
		}
	}
	return n
}

func (p *VP9FrameBufferPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.this != nil {
		gopointer.Unref(p.this)
		p.this = nil
	}

	for _, buf := range p.allocatedBuffers {
		buf.Release()
	}
	p.allocatedBuffers = nil
}

type VP9FrameBuffer struct {
	data     *C.uint8_t
	size     C.size_t
	refCount int

	mu sync.Mutex
}

func (p *VP9FrameBuffer) GetData() *C.uint8_t {
	return p.data
}

func (p *VP9FrameBuffer) GetDataSize() C.size_t {
	return p.size
}

func (p *VP9FrameBuffer) SetSize(size C.size_t) {
	if size < 1 {
		return
	}

	if p.data != nil {
		C.free(unsafe.Pointer(p.data))
		p.data = nil
	}

	p.data = C.newFrameBuffer(size)
}

func (p *VP9FrameBuffer) AddRef() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refCount++
	return p.refCount
}

func (p *VP9FrameBuffer) HasOneRef() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.refCount == 1
}

func (p *VP9FrameBuffer) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.refCount > 0 {
		p.refCount--
	}

	if p.refCount < 1 {
		C.free(unsafe.Pointer(p.data))
		p.data = nil
	}
}

//export goVpxGetFrameBuffer
func goVpxGetFrameBuffer(userPriv unsafe.Pointer, minSize C.size_t, fb *C.vpx_codec_frame_buffer_t) C.int32_t {
	fmt.Println("goVpxGetFrameBuffer")
	pool := gopointer.Restore(userPriv).(*VP9FrameBufferPool)
	buf, err := pool.GetFrameBuffer(minSize)
	if err != nil {
		return -1
	}
	buf.AddRef()
	fb.data = buf.GetData()
	fb.size = buf.GetDataSize()
	fb.priv = gopointer.Save(buf)
	return 0
}

//export goVpxReleaseFrameBuffer
func goVpxReleaseFrameBuffer(userPriv unsafe.Pointer, fb *C.vpx_codec_frame_buffer_t) C.int32_t {
	fmt.Println("goVpxReleasFrameBuffer")
	buf := gopointer.Restore(fb.priv).(*VP9FrameBuffer)
	if buf != nil {
		buf.Release()
		gopointer.Unref(fb.priv)
		fb.priv = nil
	}
	return 0
}
