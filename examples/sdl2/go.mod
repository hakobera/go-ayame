module github.com/hakobera/go-ayame/examples/sdl2

go 1.13

replace github.com/hakobera/go-ayame => ../../../go-ayame

replace github.com/hakobera/go-ayame/pkg/vpx => ../../pkg/vpx

require (
	github.com/hakobera/go-ayame v0.0.0-00010101000000-000000000000
	github.com/hakobera/go-ayame/pkg/vpx v0.0.0-00010101000000-000000000000
	github.com/pion/rtp v1.4.0
	github.com/pion/webrtc/v2 v2.2.5
	github.com/veandco/go-sdl2 v0.4.1
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1
)
