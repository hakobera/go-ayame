package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	signalingURL := flag.String("url", "wss://ayame-lite.shiguredo.jp/signaling", "Specify Ayame service address")
	roomID := flag.String("room-id", "", "specify room ID")
	signalingKey := flag.String("signaling-key", "", "specify signaling key")
	codec := flag.String("codec", "VP8", "sepcify coded (VP8 or VP9)")
	verbose := flag.Bool("verbose", false, "enable verbose log")

	flag.Parse()
	log.Printf("args: url=%s, roomID=%s, signalingKey=%s, coded=%s", *signalingURL, *roomID, *signalingKey, *codec)

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

	frameData := make(chan VpxFrame)
	defer close(frameData)

	decoder, err := NewDecoder(*codec)
	if err != nil {
		log.Printf("Failed to create VideoDecoder")
		panic(err)
	}
	defer decoder.Close()

	vpxSampleBuilder := decoder.NewSampleBuilder()

	go decoder.Process(videoData, frameData)

	opts := ayame.DefaultOptions()
	opts.SignalingKey = *signalingKey
	opts.Video.Codec = *codec
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
