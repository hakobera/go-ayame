package main

import (
	"flag"
	"fmt"
	"image"
	"log"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/hakobera/go-ayame/pkg/vpx"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/veandco/go-sdl2/sdl"

	"golang.org/x/image/draw"
)

func main() {
	signalingURL := flag.String("url", "wss://ayame-lite.shiguredo.jp/signaling", "Specify Ayame service address")
	roomID := flag.String("room-id", "", "specify room ID")
	signalingKey := flag.String("signaling-key", "", "specify signaling key")
	verbose := flag.Bool("verbose", false, "enable verbose log")

	codec := "VP8"

	const WindowWidth = 640
	const WindowHeight = 480

	flag.Parse()
	log.Printf("args: url=%s, roomID=%s, signalingKey=%s, coded=%s", *signalingURL, *roomID, *signalingKey, codec)

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Printf("Failed to initialize SDL")
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("go-ayame SDL2 example", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WindowWidth, WindowHeight, sdl.WINDOW_SHOWN)
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

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, WindowWidth, WindowHeight)
	if err != nil {
		log.Printf("Failed to create SDL texture")
		panic(err)
	}
	defer texture.Destroy()

	renderer.SetDrawColor(0, 0, 0, sdl.ALPHA_OPAQUE)
	renderer.Clear()

	decoder, err := vpx.NewDecoder(codec)
	if err != nil {
		log.Printf("Failed to create VideoDecoder")
		panic(err)
	}
	defer decoder.Close()

	vpxSampleBuilder := decoder.NewSampleBuilder()

	videoData := make(chan *media.Sample, 60)
	defer close(videoData)

	frameData := make(chan vpx.VpxFrame)

	go decoder.Process(videoData, frameData)

	opts := ayame.DefaultOptions()
	opts.SignalingKey = *signalingKey
	opts.Video.Codec = codec
	opts.Audio.Enabled = false
	con := ayame.NewConnection(*signalingURL, *roomID, opts, *verbose, false)
	defer con.Disconnect()

	con.OnConnect(func() {
		fmt.Println("Connected")
	})

	con.OnTrackPacket(func(track *webrtc.Track, packet *rtp.Packet) {
		switch track.Kind() {
		case webrtc.RTPCodecTypeVideo:
			vpxSampleBuilder.Push(packet)

			for {
				sample := vpxSampleBuilder.Pop()
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
			var err error = nil
			select {
			case f, ok := <-frameData:
				if !ok {
					return
				}

				b := f.Image.Bounds()
				if b.Dx() == WindowWidth && b.Dy() == WindowHeight {
					err = texture.Update(nil, f.Image.Pix, WindowWidth*4)
				} else {
					dst := image.NewRGBA(image.Rect(0, 0, WindowWidth, WindowHeight))
					draw.BiLinear.Scale(dst, dst.Bounds(), f.Image, f.Image.Bounds(), draw.Over, nil)
					err = texture.Update(nil, dst.Pix, WindowWidth*4)
				}

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
