package ayame_test

import (
	"reflect"
	"testing"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/pion/webrtc/v2"
)

func TestDefaultOptions(t *testing.T) {
	opts := ayame.DefaultOptions()

	if opts.Audio.Direction != webrtc.RTPTransceiverDirectionRecvonly {
		t.Errorf("Audio.Direction should be \"recvonly\"")
	}
	if !opts.Audio.Enabled {
		t.Errorf("Audio.Enabled should be true")
	}
	if opts.Audio.Bitrate != 48000 {
		t.Errorf("Audio.Bitrate should be 480000")
	}

	if opts.Video.Direction != webrtc.RTPTransceiverDirectionRecvonly {
		t.Errorf("Video.Direction should be \"recvonly\"")
	}
	if !opts.Video.Enabled {
		t.Errorf("Video.Enabled should be true")
	}
	if opts.Video.Codec != "VP8" {
		t.Errorf("Video.Codec should be \"VP8\"")
	}
	if opts.Video.Bitrate != 90000 {
		t.Errorf("Video.Bitrate should be 90000")
	}

	iceServer := opts.ICEServers[0]
	expectedIceServers := []string{"stun:stun.l.google.com:19302"}
	if !reflect.DeepEqual(iceServer.URLs, expectedIceServers) {
		t.Errorf("ICEServers[0].URLs should be %v, but %v", expectedIceServers, iceServer.URLs)
	}
	if iceServer.Username != "" {
		t.Errorf("ICEServers[0].Username should be empty")
	}
	if iceServer.Credential != nil {
		t.Errorf("ICEServers[0].Credential should be nil")
	}

	if opts.ClientID == "" {
		t.Errorf("ClientID should not be empty")
	}

	if len(opts.ICEServers) != 1 {
		t.Errorf("ICEServers should have 1 ICEServer")
	}
}

func TestNewConnection(t *testing.T) {
	signalingURL := "wss://ayame-lite.shiguredo.jp/signaling"
	roomID := "room1"
	defaultOptions := ayame.DefaultOptions()

	conn := ayame.NewConnection(signalingURL, roomID, defaultOptions, false, false)

	if conn.SignalingURL != signalingURL {
		t.Errorf("SignalingURL: got %v, but want %v", conn.SignalingURL, signalingURL)
	}

	if conn.RoomID != roomID {
		t.Errorf("RoomID: got %v, but want %v", conn.RoomID, roomID)
	}

	if !reflect.DeepEqual(conn.Options, defaultOptions) {
		t.Errorf("Options: got %v, but want %v", conn.Options, defaultOptions)
	}

	if conn.Debug {
		t.Errorf("Debug: got true, but want false")
	}
}
