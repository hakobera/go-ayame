module github.com/hakobera/go-ayame/examples/sdl2

go 1.13

replace github.com/hakobera/go-ayame v0.2.0 => ../../../go-ayame

replace github.com/hakobera/go-ayame/pkg/vpx v0.2.0 => ../../pkg/vpx

require (
	github.com/hakobera/go-ayame v0.2.0
	github.com/hakobera/go-ayame/pkg/vpx v0.2.0
	github.com/pion/rtp v1.4.0
	github.com/pion/webrtc/v2 v2.2.9
	github.com/veandco/go-sdl2 v0.4.3
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1
)
