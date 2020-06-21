package decoder

import (
	"image"
)

// DecodedImage defines interface for images which decoded by encoded video frame.
type DecodedImage interface {
	IsKeyFrame() bool
	Width() uint32
	Height() uint32
	Plane(n int) []byte
	Stride(n int) int
	ToBytes(format ColorFormat) []byte
	ToRGBA() *image.RGBA
}
