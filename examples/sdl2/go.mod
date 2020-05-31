module github.com/hakobera/go-ayame/examples/sdl2

go 1.13

replace github.com/hakobera/go-ayame v0.2.0 => ../../../go-ayame

replace github.com/hakobera/go-ayame/pkg/decoder v0.2.1 => ../../pkg/decoder

require (
	github.com/hakobera/go-ayame v0.2.0
	github.com/hakobera/go-ayame/pkg/decoder v0.2.1
	github.com/pion/rtp v1.5.2
	github.com/pion/webrtc/v2 v2.2.11
	github.com/veandco/go-sdl2 v0.4.3
)
