module github.com/hakobera/go-ayame/examples/save-to-webm

go 1.13

replace github.com/hakobera/go-ayame v0.2.0 => ../../../go-ayame

require (
	github.com/at-wat/ebml-go v0.11.0
	github.com/hakobera/go-ayame v0.2.0
	github.com/pion/rtp v1.5.5
	github.com/pion/webrtc/v2 v2.2.16
)
