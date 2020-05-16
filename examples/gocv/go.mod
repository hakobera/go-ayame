module github.com/hakobera/go-ayame/examples/gocv

go 1.14

replace github.com/hakobera/go-ayame v0.2.0 => ../../../go-ayame

replace github.com/hakobera/go-ayame/pkg/vpx v0.2.0 => ../../pkg/vpx

require (
	github.com/hakobera/go-ayame v0.2.0
	github.com/hakobera/go-ayame/pkg/vpx v0.2.0
	github.com/pion/rtp v1.5.2
	github.com/pion/webrtc/v2 v2.2.11
	gocv.io/x/gocv v0.22.0
)
