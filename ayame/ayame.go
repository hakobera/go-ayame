// Package ayame は Ayame クライアントライブラリです。
package ayame

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

// DefaultOptions は Ayame 接続オプションのデフォルト値を生成して返します。
func DefaultOptions() *ConnectionOptions {
	return &ConnectionOptions{
		Audio: ConnectionAudioOption{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
			Enabled:   true,
			Bitrate:   48000,
		},
		Video: ConnectionVideoOption{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
			Enabled:   true,
			Codec:     "VP8",
			Bitrate:   90000,
		},
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		ClientID: getULID(),
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
		onTrackPacketHandler: func(track *webrtc.Track, packet *rtp.Packet) {},
		onByeHandler:         func() {},
		onDataChannelHandler: func(dc *webrtc.DataChannel) {},
	}

	return c
}
