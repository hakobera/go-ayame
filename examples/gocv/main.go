package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/hakobera/go-ayame/pkg/decoder"
	"github.com/hakobera/go-ayame/pkg/decoder/vpx"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
	"gocv.io/x/gocv"
)

const (
	frameX      = 320
	frameY      = 240
	minimumArea = 3000
)

var dc *webrtc.DataChannel = nil

func main() {
	signalingURL := flag.String("url", "wss://ayame-lite.shiguredo.jp/signaling", "Specify Ayame service address")
	roomID := flag.String("room-id", "", "specify room ID")
	signalingKey := flag.String("signaling-key", "", "specify signaling key")
	verbose := flag.Bool("verbose", false, "enable verbose log")

	flag.Parse()
	log.Printf("args: url=%s, roomID=%s, signalingKey=%s", *signalingURL, *roomID, *signalingKey)

	opts := ayame.DefaultOptions()
	opts.SignalingKey = *signalingKey
	opts.Audio.Enabled = false

	d, err := vpx.NewVP8Decoder()
	if err != nil {
		log.Printf("Failed to create VideoDecoder")
		panic(err)
	}
	defer d.Close()

	videoFrameBuilder := d.NewFrameBuilder()

	videoData := make(chan *decoder.Frame, 60)
	defer close(videoData)

	imgData := make(chan vpx.DecodedImage)

	go d.Process(videoData, imgData)

	con := ayame.NewConnection(*signalingURL, *roomID, opts, *verbose, false)
	defer con.Disconnect()

	con.OnConnect(func() {
		fmt.Println("Connected")

		var err error = nil
		dc, err = con.CreateDataChannel("dataChannel1", nil)
		if err != nil {
			log.Printf("CreateDataChannel error: %v", err)
			return
		}
		dc.OnMessage(onMessage(dc))
	})

	con.OnTrackPacket(func(track *webrtc.Track, packet *rtp.Packet) {
		switch track.Kind() {
		case webrtc.RTPCodecTypeVideo:
			videoFrameBuilder.Push(packet)

			for {
				frame := videoFrameBuilder.Pop()
				if frame == nil {
					return
				}
				videoData <- frame
			}
		}
	})

	con.OnDataChannel(func(c *webrtc.DataChannel) {
		if dc == nil {
			dc = c
			dc.OnMessage(onMessage(dc))
		}
	})

	err = con.Connect()
	if err != nil {
		log.Fatal("failed to connect Ayame", err)
	}

	startGoCVMotionDetect(imgData)
}

// This was taken from the GoCV examples, the only change is we are taking a buffer from WebRTC instead of webcam
// https://github.com/hybridgroup/gocv/blob/master/cmd/motion-detect/main.go
func startGoCVMotionDetect(imgData <-chan vpx.DecodedImage) {
	window := gocv.NewWindow("Motion Window")
	defer window.Close() //nolint

	img := gocv.NewMat()
	defer img.Close() //nolint

	imgDelta := gocv.NewMat()
	defer imgDelta.Close() //nolint

	imgThresh := gocv.NewMat()
	defer imgThresh.Close() //nolint

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close() //nolint

	prevStatus := "Ready"

L:
	for {
		select {
		case img, ok := <-imgData:
			if !ok {
				break L
			}

			buf := img.ToBytes(decoder.ColorBGRA)
			mat, _ := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC4, buf)
			if mat.Empty() {
				continue
			}

			status := "Ready"
			statusColor := color.RGBA{0, 255, 0, 0}

			// first phase of cleaning up image, obtain foreground only
			mog2.Apply(mat, &imgDelta)

			// remaining cleanup of the image to use for finding contours.
			// first use threshold
			gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

			// then dilate
			kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
			defer kernel.Close() //nolint
			gocv.Dilate(imgThresh, &imgThresh, kernel)

			// now find contours
			contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
			for i, c := range contours {
				area := gocv.ContourArea(c)
				if area < minimumArea {
					continue
				}

				status = "Motion detected"
				statusColor = color.RGBA{255, 0, 0, 0}
				gocv.DrawContours(&mat, contours, i, statusColor, 2)

				rect := gocv.BoundingRect(c)
				gocv.Rectangle(&mat, rect, color.RGBA{0, 0, 255, 0}, 2)
			}

			gocv.PutText(&mat, status, image.Pt(10, 30), gocv.FontHersheyPlain, 2.0, statusColor, 2)

			if prevStatus != status {
				if status == "Motion detected" {
					dc.SendText("o")
				} else {
					dc.SendText("p")
				}
			}
			prevStatus = status

			window.IMShow(mat)
			if window.WaitKey(1) == 27 {
				break L
			}
		}
	}
}

func onMessage(dc *webrtc.DataChannel) func(webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			fmt.Printf("OnData[%s]: data=%s\n", dc.Label(), (msg.Data))
		}
	}
}
