module github.com/hakobera/go-ayame/examples/gocv

go 1.14

replace github.com/hakobera/go-ayame v0.2.0 => ../../../go-ayame

require (
	github.com/hakobera/go-ayame v0.2.0
	github.com/hakobera/go-webrtc-decoder v0.1.0
	github.com/pion/rtp v1.5.5
	github.com/pion/webrtc/v2 v2.2.17
	gocv.io/x/gocv v0.23.0
)
