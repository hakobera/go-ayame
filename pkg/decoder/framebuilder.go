package decoder

import (
	"reflect"

	"github.com/pion/rtp"
)

type Frame struct {
	Packets   []interface{}
	Timestamp uint32
}

// FrameBuilder contains all packets
// maxLate determines how long we should wait until we get a valid Frame
// The larger the value the less packet loss you will see, but higher latency
//
// This is customized version of Pion's sample builder
// https://github.com/pion/webrtc/blob/master/pkg/media/samplebuilder/samplebuilder.go
type FrameBuilder struct {
	maxLate uint16
	buffer  [65536]*rtp.Packet

	// Interface that allows us to take RTP packets to samples
	depacketizer rtp.Depacketizer

	// Last seqnum that has been added to buffer
	lastPush uint16

	// Last seqnum that has been successfully popped
	// isContiguous is false when we start or when we have a gap
	// that is older then maxLate
	isContiguous     bool
	lastPopSeq       uint16
	lastPopTimestamp uint32

	// Interface that checks whether the packet is the first fragment of the frame or not
	partitionHeadChecker rtp.PartitionHeadChecker
}

// NewFrameBuilder constructs a new FrameBuilder
func NewFrameBuilder(maxLate uint16, depacketizer rtp.Depacketizer, checker rtp.PartitionHeadChecker, opts ...Option) *FrameBuilder {
	s := &FrameBuilder{maxLate: maxLate, depacketizer: depacketizer, partitionHeadChecker: checker}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Push adds a RTP Packet to the FrameBuilder
func (s *FrameBuilder) Push(p *rtp.Packet) {
	s.buffer[p.SequenceNumber] = p
	s.lastPush = p.SequenceNumber
	s.buffer[p.SequenceNumber-s.maxLate] = nil
}

// We have a valid collection of RTP Packets
// walk forwards building a sample if everything looks good clear and update buffer+values
func (s *FrameBuilder) buildFrame(firstBuffer uint16) *Frame {
	packets := []interface{}{}

	for i := firstBuffer; s.buffer[i] != nil; i++ {
		if s.buffer[i].Timestamp != s.buffer[firstBuffer].Timestamp {
			s.lastPopSeq = i - 1
			s.isContiguous = true
			s.lastPopTimestamp = s.buffer[i-1].Timestamp
			for j := firstBuffer; j < i; j++ {
				s.buffer[j] = nil
			}
			return &Frame{Packets: packets, Timestamp: s.lastPopTimestamp}
		}

		_, err := s.depacketizer.Unmarshal(s.buffer[i].Payload)
		if err != nil {
			return nil
		}

		packets = append(packets, reflect.Indirect(reflect.ValueOf(s.depacketizer)).Interface())
	}
	return nil
}

// Distance between two seqnums
func seqnumDistance(x, y uint16) uint16 {
	diff := int16(x - y)
	if diff < 0 {
		return uint16(-diff)
	}

	return uint16(diff)
}

// Pop scans buffer for valid samples, returns nil when no valid samples have been found
func (s *FrameBuilder) Pop() *Frame {
	return s.PopWithTimestamp()
}

// PopWithTimestamp scans buffer for valid samples and its RTP timestamp,
// returns nil, 0 when no valid samples have been found
func (s *FrameBuilder) PopWithTimestamp() *Frame {
	var i uint16
	if !s.isContiguous {
		i = s.lastPush - s.maxLate
	} else {
		if seqnumDistance(s.lastPopSeq, s.lastPush) > s.maxLate {
			i = s.lastPush - s.maxLate
			s.isContiguous = false
		} else {
			i = s.lastPopSeq + 1
		}
	}

	for ; i != s.lastPush; i++ {
		curr := s.buffer[i]
		if curr == nil {
			continue // we haven't hit a buffer yet, keep moving
		}

		if !s.isContiguous {
			if !s.partitionHeadChecker.IsPartitionHead(curr.Payload) {
				continue
			}
			// We can start using this frame as it is a head of frame partition
		}

		// Initial validity checks have passed, walk forward
		return s.buildFrame(i)
	}
	return nil
}

// Option configures FrameBuilder
type Option func(o *FrameBuilder)
