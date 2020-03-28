module github.com/hakobera/go-ayame/examples/gocv

go 1.14

replace github.com/hakobera/go-ayame => ../../../go-ayame

replace github.com/hakobera/go-ayame/pkg/vpx => ../../pkg/vpx

require (
	github.com/hakobera/go-ayame v0.0.0-00010101000000-000000000000
	github.com/hakobera/go-ayame/pkg/vpx v0.0.0-00010101000000-000000000000
	github.com/pion/rtp v1.4.0
	github.com/pion/webrtc/v2 v2.2.4
	gocv.io/x/gocv v0.22.0
)
