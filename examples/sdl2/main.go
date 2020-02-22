package main

import (
	"flag"
	"fmt"
	"image"
	"log"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/hakobera/libvpx-go/vpx"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	signalingURL := flag.String("url", "wss://ayame-lite.shiguredo.jp/signaling", "Specify Ayame service address")
	roomID := flag.String("room-id", "", "specify room ID")
	signalingKey := flag.String("signaling-key", "", "specify signaling key")
	codec := flag.String("codec", "VP8", "sepcify coded (VP8 or VP9)")
	verbose := flag.Bool("verbose", false, "enable verbose log")

	flag.Parse()
	log.Printf("args: url=%s, roomID=%s, signalingKey=%s", *signalingURL, *roomID, *signalingKey)

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Printf("Failed to initialize SDL")
		panic(err)
	}
	defer sdl.Quit()

	var width int32 = 640
	var height int32 = 480
	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Printf("Failed to create SDL window")
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		log.Printf("Failed to create SDL renderer")
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, width, height)
	if err != nil {
		log.Printf("Failed to create SDL texture")
		panic(err)
	}
	defer texture.Destroy()

	renderer.SetDrawColor(0, 0, 0, sdl.ALPHA_OPAQUE)
	renderer.Clear()

	videoData := make(chan *media.Sample, 60)
	defer close(videoData)

	videoBuilder := samplebuilder.New(10, &codecs.VP8Packet{})

	frameData := make(chan frame)
	decoder, err := newVideoDecoder(*codec, videoData)
	if err != nil {
		log.Printf("Failed to create VideoDecoder")
		panic(err)
	}

	go decoder.Process(frameData)

	opts := ayame.DefaultOptions()
	opts.SignalingKey = *signalingKey
	opts.Audio.Enabled = false
	con := ayame.NewConnection(*signalingURL, *roomID, opts, *verbose, false)
	defer con.Disconnect()

	con.OnConnect(func() {
		fmt.Println("Connected")
	})

	con.OnTrackPacket(func(track *webrtc.Track, packet *rtp.Packet) {
		switch track.Kind() {
		case webrtc.RTPCodecTypeVideo:
			videoBuilder.Push(packet)

			for {
				sample := videoBuilder.Pop()
				if sample == nil {
					return
				}
				videoData <- sample
			}
		}
	})

	err = con.Connect()
	if err != nil {
		log.Fatal("failed to connect Ayame", err)
	}

	go func() {
		for {
			select {
			case f, ok := <-frameData:
				if !ok {
					return
				}
				err := texture.Update(nil, f.RGBA.Pix, 640*4)
				if err != nil {
					log.Println("Failed to update SDL Texture", err)
					continue
				}
				window.UpdateSurface()

				renderer.Clear()
				renderer.Copy(texture, nil, nil)
				renderer.Present()
			}
		}
	}()

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

type frame struct {
	*image.RGBA
	IsKeyframe bool
	Width      int
	Height     int
}

type videoDecoder struct {
	src   <-chan *media.Sample
	ctx   *vpx.CodecCtx
	iface *vpx.CodecIface
}

func newVideoDecoder(codec string, src chan *media.Sample) (*videoDecoder, error) {
	dec := &videoDecoder{
		src: src,
		ctx: vpx.NewCodecCtx(),
	}
	switch codec {
	case "VP8":
		dec.iface = vpx.DecoderIfaceVP8()
	case "VP9":
		dec.iface = vpx.DecoderIfaceVP9()
	default:
		return nil, fmt.Errorf("Unsupported coded: %s", codec)
	}
	err := vpx.Error(vpx.CodecDecInitVer(dec.ctx, dec.iface, nil, 0, vpx.DecoderABIVersion))
	if err != nil {
		return nil, err
	}
	return dec, nil
}

func (v *videoDecoder) Process(out chan<- frame) {
	defer close(out)
	for pkt := range v.src {
		dataSize := uint32(len(pkt.Data))
		err := vpx.Error(vpx.CodecDecode(v.ctx, string(pkt.Data), dataSize, nil, 0))
		if err != nil {
			log.Println("[WARN]", err)
			continue
		}
		isKeyframe := (pkt.Data[0]&0x1 == 0)
		width := 0
		height := 0
		if isKeyframe {
			raw := uint(pkt.Data[6]) | uint(pkt.Data[7])<<8 | uint(pkt.Data[8])<<16 | uint(pkt.Data[9])<<24
			width = int(raw & 0x3FFF)
			height = int((raw >> 16) & 0x3FFF)
		}
		if width != 0 {
			log.Printf("width=%d, height=%d\n", width, height)
		}

		var iter *vpx.CodecIter
		img := vpx.CodecGetFrame(v.ctx, &iter)
		for img != nil {
			img.Deref()
			out <- frame{
				RGBA:       img.ImageRGBA(),
				IsKeyframe: (pkt.Data[0]&0x1 == 0),
				Width:      width,
				Height:     height,
			}
			img = vpx.CodecGetFrame(v.ctx, &iter)
		}
	}
}
