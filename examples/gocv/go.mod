module github.com/hakobera/go-ayame/examples/gocv

go 1.14

replace github.com/hakobera/go-ayame v0.2.0 => ../../../go-ayame

replace github.com/hakobera/go-ayame/pkg/decoder v0.2.1 => ../../pkg/decoder

require (
	github.com/hakobera/go-ayame v0.2.0
	github.com/hakobera/go-ayame/pkg/decoder v0.2.1
	github.com/pion/rtp v1.5.2
	github.com/pion/webrtc/v2 v2.2.11
	gocv.io/x/gocv v0.23.0
)
