package decoder

type Decoder interface {
	NewFrameBuilder() *FrameBuilder
	Process(src <-chan *Frame, out chan<- DecodedImage)
	Close() error
}
