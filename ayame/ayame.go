// Package ayame は Ayame クライアントライブラリです。
package ayame

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

// DefaultOptions は Ayame 接続オプションのデフォルト値を生成して返します。
func DefaultOptions() *ConnectionOptions {
	return &ConnectionOptions{
		Audio: ConnectionAudioOption{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
			Enabled:   false,
			Codecs: []*webrtc.RTPCodecParameters{
				{
					// RFC 7587 "a=rtpmap" MUST be 48000, and the number of channels MUST be 2.
					RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2, SDPFmtpLine: "", RTCPFeedback: nil},
					PayloadType:        111,
				},
			},
		},
		Video: ConnectionVideoOption{
			Direction: webrtc.RTPTransceiverDirectionSendonly,
			Enabled:   true,
			Codecs: []*webrtc.RTPCodecParameters{
				{
					RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
					PayloadType:        96,
				},
			},
		},
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
		ClientID:     getULID(),
		UseTrickeICE: true,
	}
}

// NewConnection は Ayame Connection を生成して返します。
func NewConnection(signalingURL string, roomID string, options *ConnectionOptions, debug bool, isRelay bool) *Connection {
	transportPolicy := webrtc.ICETransportPolicyAll
	if isRelay {
		transportPolicy = webrtc.ICETransportPolicyRelay
	}

	if options == nil {
		options = DefaultOptions()
	}

	c := &Connection{
		SignalingURL:  signalingURL,
		RoomID:        roomID,
		Options:       options,
		Debug:         debug,
		AuthnMetadata: nil,

		authzMetadata:   nil,
		connectionState: webrtc.ICEConnectionStateNew,
		connectionID:    "",
		ws:              nil,
		pc:              nil,
		pcConfig: webrtc.Configuration{
			ICEServers:         options.ICEServers,
			ICETransportPolicy: transportPolicy,
		},
		isOffer:       false,
		isExistClient: false,

		dataChannels: map[string]*webrtc.DataChannel{},

		onOpenHandler:        func(metadata *interface{}) {},
		onConnectHandler:     func() {},
		onDisconnectHandler:  func(reason string, err error) {},
		onTrackPacketHandler: func(track *webrtc.TrackRemote, packet *rtp.Packet) {},
		onByeHandler:         func() {},
		onDataChannelHandler: func(dc *webrtc.DataChannel) {},
	}

	return c
}
