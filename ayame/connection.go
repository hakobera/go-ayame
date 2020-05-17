package ayame

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
)

const (
	readTimeout  = 90 * time.Second
	readLimit    = 1048576
	writeTimeout = 10 * time.Second
)

// Connection は PeerConnection 接続を管理します。
type Connection struct {
	// シグナリングに利用する URL
	SignalingURL string

	// Ayame のルームID
	RoomID string

	// Ayame の接続オプション
	Options *ConnectionOptions

	// デバッグログの出力の可否
	Debug bool

	// 送信する認証用のメタデータ
	AuthnMetadata *interface{}

	// MediaStream API is not yet fully supported by pion.
	// Only working on Linux machine
	// Check development status of https://github.com/pion/mediadevices
	//Stream          *msapi.MediaStream
	//RemoteStream    *msapi.MediaStream

	authzMetadata   *interface{}
	connectionState webrtc.ICEConnectionState
	connectionID    string

	ws            *websocket.Conn
	pc            *webrtc.PeerConnection
	pcConfig      webrtc.Configuration
	isOffer       bool
	isExistClient bool

	dataChannels map[string]*webrtc.DataChannel

	onOpenHandler        func(metadata *interface{})
	onConnectHandler     func()
	onDisconnectHandler  func(reason string, err error)
	onTrackPacketHandler func(track *webrtc.Track, packet *rtp.Packet)
	onByeHandler         func()
	onDataChannelHandler func(dc *webrtc.DataChannel)

	callbackMu sync.Mutex
}

// Connect は PeerConnection 接続を開始します。
func (c *Connection) Connect() error {
	if c.ws != nil || c.pc != nil {
		c.trace("connection already exists")
		return fmt.Errorf("connection alreay exists")
	}
	c.signaling()
	return nil
}

// Disconnect は PeerConnection 接続を切断します。
func (c *Connection) Disconnect() {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()

	for _, dc := range c.dataChannels {
		c.closeDataChannel(dc)
	}

	c.closePeerConnection()
	c.closeWebSocketConnection()
	c.authzMetadata = nil
	c.connectionID = ""
	c.connectionState = webrtc.ICEConnectionStateNew
	c.isOffer = false
	c.isExistClient = false

	c.dataChannels = map[string]*webrtc.DataChannel{}

	c.onOpenHandler = func(metadata *interface{}) {}
	c.onConnectHandler = func() {}
	c.onDisconnectHandler = func(reason string, err error) {}
	c.onTrackPacketHandler = func(track *webrtc.Track, packet *rtp.Packet) {}
	c.onByeHandler = func() {}
	c.onDataChannelHandler = func(dc *webrtc.DataChannel) {}
}

// CreateDataChannel は指定した label と options から新しい DataChannel 作成して、追加します。
func (c *Connection) CreateDataChannel(label string, options *webrtc.DataChannelInit) (*webrtc.DataChannel, error) {
	if c.pc == nil {
		return nil, fmt.Errorf("PeerConnection Does Not Ready")
	}
	if c.isOffer {
		return nil, fmt.Errorf("PeerConnection Has Local Offer")
	}
	if _, ok := c.dataChannels[label]; ok {
		return nil, fmt.Errorf("DataChannel Already Exists. label=%s", label)
	}

	if c.isExistClient {
		dc, err := c.pc.CreateDataChannel(label, options)
		if err != nil {
			return nil, err
		}

		dc.OnOpen(func() {
			c.trace("datachannel OnOpen")
		})
		dc.OnClose(func() {
			c.trace("datachannel OnClose")
			delete(c.dataChannels, label)
		})
		dc.OnError(func(err error) {
			c.trace("datachannel OnError: %v", err)
			delete(c.dataChannels, label)
		})
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			c.trace("datachannel OnMessage")
		})

		c.dataChannels[label] = dc
		return dc, nil
	}
	return nil, fmt.Errorf("client does not exist")
}

// OnOpen は open イベント発生時のコールバック関数を設定します。
func (c *Connection) OnOpen(f func(metadata *interface{})) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onOpenHandler = f
}

// OnConnect は connect イベント発生時のコールバック関数を設定します。
func (c *Connection) OnConnect(f func()) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onConnectHandler = f
}

// OnDisconnect は disconnect イベント発生時のコールバック関数を設定します。
func (c *Connection) OnDisconnect(f func(reason string, err error)) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onDisconnectHandler = f
}

// OnTrackPacket は RTP Packet 受診時に発生するコールバック関数を設定します。
func (c *Connection) OnTrackPacket(f func(track *webrtc.Track, packet *rtp.Packet)) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onTrackPacketHandler = f
}

// OnBye は bye イベント発生時のコールバック関数を設定します。
func (c *Connection) OnBye(f func()) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onByeHandler = f
}

// OnDataChannel は datachannel イベント発生時のコールバック関数を設定します。
func (c *Connection) OnDataChannel(f func(dc *webrtc.DataChannel)) {
	c.callbackMu.Lock()
	defer c.callbackMu.Unlock()
	c.onDataChannelHandler = f
}

func (c *Connection) trace(format string, v ...interface{}) {
	if c.Debug {
		logf(format, v...)
	}
}

func (c *Connection) signaling() error {
	if c.ws != nil {
		return fmt.Errorf("WS-ALREADY-EXISTS")
	}

	ctx := context.Background()

	ws, err := c.openWS(ctx)
	if err != nil {
		return fmt.Errorf("WS-OPEN-ERROR: %w", err)
	}
	c.ws = ws

	ctx, cancel := context.WithCancel(ctx)
	messageChannel := make(chan []byte, 100)

	go c.recv(ctx, messageChannel)
	go c.main(cancel, messageChannel)

	return c.sendRegisterMessage()
}

func (c *Connection) openWS(ctx context.Context) (*websocket.Conn, error) {
	c.trace("Connecting to %s", c.SignalingURL)
	conn, _, err := websocket.Dial(ctx, c.SignalingURL, nil)
	if err != nil {
		return nil, err
	}
	c.trace("Connected to %s", c.SignalingURL)
	return conn, nil
}

func (c *Connection) sendMsg(v interface{}) error {
	if c.ws != nil {
		ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()
		c.trace("send %v", v)
		if err := wsjson.Write(ctx, c.ws, v); err != nil {
			c.trace("failed to send %v: %v", v, err)
			return err
		}
	}
	return nil
}

func (c *Connection) sendPongMessage() error {
	msg := &pongMessage{
		Type: "pong",
	}

	if err := c.sendMsg(msg); err != nil {
		return err
	}
	return nil
}

func (c *Connection) sendRegisterMessage() error {
	msg := &registerMessage{
		Type:          "register",
		RoomID:        c.RoomID,
		ClientID:      c.Options.ClientID,
		AuthnMetadata: nil,
		SignalingKey:  nil,

		AyameClient: strPtr("go-ayame v0.1.0"),
		Environment: strPtr(runtime.GOOS + " " + runtime.GOARCH),
	}
	if c.AuthnMetadata != nil {
		msg.AuthnMetadata = c.AuthnMetadata
	}
	if c.Options.SignalingKey != "" {
		msg.SignalingKey = &c.Options.SignalingKey
	}

	if err := c.sendMsg(msg); err != nil {
		return err
	}
	return nil
}

func (c *Connection) sendSdp(sessionDescription *webrtc.SessionDescription) {
	c.sendMsg(sessionDescription)
}

func (c *Connection) createPeerConnection() error {
	m := webrtc.MediaEngine{}
	if c.Options.Audio.Enabled {
		for _, codec := range c.Options.Audio.Codecs {
			m.RegisterCodec(codec)
		}
	}
	if c.Options.Video.Enabled {
		for _, codec := range c.Options.Video.Codecs {
			m.RegisterCodec(codec)
		}
	}

	s := webrtc.SettingEngine{}
	s.SetTrickle(c.Options.UseTrickeICE)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithSettingEngine(s))

	c.trace("RTCConfiguration: %v", c.pcConfig)
	pc, err := api.NewPeerConnection(c.pcConfig)
	if err != nil {
		return err
	}

	if c.Options.Audio.Enabled {
		_, err = pc.AddTransceiver(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{
			Direction: c.Options.Audio.Direction,
		})
		if err != nil {
			return err
		}
	}

	if c.Options.Video.Enabled {
		_, err = pc.AddTransceiver(webrtc.RTPCodecTypeVideo, webrtc.RtpTransceiverInit{
			Direction: c.Options.Video.Direction,
		})
		if err != nil {
			return err
		}
	}

	// Set a Handler for when a new remote track starts, this Handler copies inbound RTP packets,
	// replaces the SSRC and sends them back
	pc.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This is a temporary fix until we implement incoming RTCP events, then we would push a PLI only when a viewer requests it
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				if c.pc == nil || c.pc.SignalingState() == webrtc.SignalingStateClosed {
					return
				}

				errSend := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}})
				if errSend != nil {
					c.trace("Failed to write RTCP packet: %s", errSend.Error())
				}
			}
		}()

		c.trace("peerConnection.ontrack(): %d, codec: %s", track.PayloadType(), track.Codec().Name)
		go func() {
			for {
				rtp, readErr := track.ReadRTP()
				if readErr != nil {
					if readErr == io.EOF {
						return
					}
					c.trace("read RTP error %v", readErr)
					c.Disconnect()
					c.onDisconnectHandler("READ-RTP-ERROR", err)
					return
				}
				c.onTrackPacketHandler(track, rtp)

				if c.pc == nil || c.pc.SignalingState() == webrtc.SignalingStateClosed {
					return
				}
			}
		}()
	})
	// Set the Handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		c.trace("ICE connection Status has changed to %s", connectionState.String())
		if c.connectionState != connectionState {
			c.connectionState = connectionState
			switch c.connectionState {
			case webrtc.ICEConnectionStateConnected:
				c.isOffer = false
				c.onConnectHandler()
			case webrtc.ICEConnectionStateDisconnected:
				fallthrough
			case webrtc.ICEConnectionStateFailed:
				c.Disconnect()
				c.onDisconnectHandler("ICE-CONNECTION-STATE-FAILED", nil)
			}
		}
	})
	// Set the Handler for Signaling connection state
	pc.OnSignalingStateChange(func(signalingState webrtc.SignalingState) {
		c.trace("signaling state changes: %s", signalingState.String())
	})

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		c.onDataChannel(dc)
	})

	if c.pc == nil {
		c.pc = pc
		c.onOpenHandler(c.authzMetadata)
	} else {
		c.pc = pc
	}
	return nil
}

func (c *Connection) sendOffer() error {
	if c.pc == nil {
		return nil
	}

	offer, err := c.pc.CreateOffer(nil)
	if err != nil {
		return err
	}
	c.trace("create offer sdp=%s", offer.SDP)
	c.pc.SetLocalDescription(offer)
	if c.pc.LocalDescription() != nil {
		c.sendSdp(c.pc.LocalDescription())
	}
	c.isOffer = true
	return nil
}

func (c *Connection) createAnswer() error {
	if c.pc == nil {
		return nil
	}

	answer, err := c.pc.CreateAnswer(nil)
	if err != nil {
		c.Disconnect()
		c.onDisconnectHandler("CREATE-ANSWER-ERROR", err)
		return err
	}
	c.trace("create answer sdp=%s", answer.SDP)
	c.pc.SetLocalDescription(answer)
	if c.pc.LocalDescription() != nil {
		c.sendSdp(c.pc.LocalDescription())
	}
	return nil
}

func (c *Connection) setAnswer(sessionDescription webrtc.SessionDescription) error {
	if c.pc == nil {
		return nil
	}
	err := c.pc.SetRemoteDescription(sessionDescription)
	if err != nil {
		return err
	}
	c.trace("set answer sdp=%s", sessionDescription.SDP)
	return nil
}

func (c *Connection) setOffer(sessionDescription webrtc.SessionDescription) error {
	if c.pc == nil {
		return nil
	}
	err := c.pc.SetRemoteDescription(sessionDescription)
	if err != nil {
		c.Disconnect()
		c.onDisconnectHandler("CREATE-OFFER-ERROR", err)
		return err
	}
	c.trace("set offer sdp=%s", sessionDescription.SDP)
	err = c.createAnswer()
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) addICECandidate(candidate webrtc.ICECandidateInit) {
	if c.pc == nil {
		return
	}
	err := c.pc.AddICECandidate(candidate)
	if err != nil {
		// ignore error
		c.trace("invalid ice candidate, %v", candidate)
		return
	}
}

func (c *Connection) onDataChannel(dc *webrtc.DataChannel) {
	label := dc.Label()
	c.trace("on data channel, label='%s'", label)
	if c.pc == nil {
		return
	}
	if dc == nil {
		return
	}
	if label == "" {
		return
	}

	dc.OnOpen(func() {
		c.trace("datachannel OnOpen")
	})
	dc.OnClose(func() {
		c.trace("datachannel OnClose")
	})
	dc.OnError(func(err error) {
		c.trace("datachannel OnError: %v", err)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		c.trace("datachannel OnMessage")
	})

	if _, ok := c.dataChannels[label]; !ok {
		c.dataChannels[label] = dc
	}

	c.onDataChannelHandler(dc)
}

func (c *Connection) closeDataChannel(dc *webrtc.DataChannel) {
	if dc.ReadyState() == webrtc.DataChannelStateClosed {
		return
	}
	dc.OnClose(func() {})
	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		for range ticker.C {
			if dc.ReadyState() == webrtc.DataChannelStateClosed {
				ticker.Stop()
				return
			}
		}
	}()
	dc.Close()
}

func (c *Connection) closePeerConnection() {
	if c.pc == nil {
		return
	}
	if c.pc != nil && c.pc.SignalingState() == webrtc.SignalingStateClosed {
		c.pc = nil
		return
	}
	c.pc.OnICEConnectionStateChange(func(_ webrtc.ICEConnectionState) {})

	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		for range ticker.C {
			if c.pc == nil {
				ticker.Stop()
				return
			}
			if c.pc != nil && c.pc.SignalingState() == webrtc.SignalingStateClosed {
				ticker.Stop()
				c.pc = nil
				return
			}
		}
	}()
	if c.pc != nil {
		c.pc.Close()
	}
}

func (c *Connection) closeWebSocketConnection() {
	if c.ws == nil {
		return
	}

	if err := c.ws.Close(websocket.StatusNormalClosure, ""); err != nil {
		c.trace("FAILED-SEND-CLOSE-MESSAGE")
	}
	c.trace("SENT-CLOSE-MESSAGE")
	c.ws = nil
}

func (c *Connection) main(cancel context.CancelFunc, messageChannel chan []byte) {
	defer func() {
		cancel()
		c.trace("EXIT-MAIN")
	}()

loop:
	for {
		select {
		case rawMessage, ok := <-messageChannel:
			if !ok {
				c.trace("CLOSED-MESSAGE-CHANNEL")
				return
			}
			if err := c.handleMessage(rawMessage); err != nil {
				break loop
			}
		}
	}
}

func (c *Connection) recv(ctx context.Context, messageChannel chan []byte) {
loop:
	for {
		if c.ws == nil {
			break loop
		}

		cctx, cancel := context.WithTimeout(ctx, readTimeout)
		_, rawMessage, err := c.ws.Read(cctx)
		cancel()
		if err != nil {
			c.trace("failed to ReadMessage: %v", err)
			break loop
		}
		messageChannel <- rawMessage
	}
	close(messageChannel)
	c.trace("CLOSE-MESSAGE-CHANNEL")
	<-ctx.Done()
	c.trace("EXITED-MAIN")
	c.Disconnect()
	c.onDisconnectHandler("EXIT-RECV", nil)
	c.trace("EXIT-RECV")
}

func (c *Connection) handleMessage(rawMessage []byte) error {
	message := &message{}
	if err := json.Unmarshal(rawMessage, &message); err != nil {
		c.trace("invalid JSON, rawMessage: %s, error: %v", rawMessage, err)
		return errorInvalidJSON
	}

	c.trace("recv type: %s, rawMessage: %s", message.Type, string(rawMessage))

	switch message.Type {
	case "ping":
		c.sendPongMessage()
	case "bye":
		c.onByeHandler()
	case "accept":
		acceptMsg := acceptMessage{}
		if err := unmarshalMessage(c, rawMessage, &acceptMsg); err != nil {
			return err
		}
		c.connectionID = acceptMsg.ConnectionID
		c.authzMetadata = acceptMsg.AuthzMetadata
		if acceptMsg.IceServers != nil && len(*acceptMsg.IceServers) != 0 {
			c.trace("IceServers: %v", *acceptMsg.IceServers)
			iceServers := make([]webrtc.ICEServer, len(*acceptMsg.IceServers))
			for i, s := range *acceptMsg.IceServers {
				iceServers[i] = webrtc.ICEServer{
					URLs: s.Urls,
				}
				if s.UserName != nil {
					iceServers[i].Username = *s.UserName
				}
				if s.Credential != nil {
					iceServers[i].Credential = *s.Credential
					iceServers[i].CredentialType = webrtc.ICECredentialTypePassword
				}
			}
			c.pcConfig.ICEServers = iceServers
		}
		c.trace("isExistClient: %v", acceptMsg.IsExistClient)
		c.isExistClient = acceptMsg.IsExistClient
		c.createPeerConnection()
		if c.isExistClient {
			return c.sendOffer()
		}
	case "reject":
		rejectMsg := rejectMessage{}
		if err := unmarshalMessage(c, rawMessage, &rejectMsg); err != nil {
			return err
		}
		c.trace("rejected, reason: %s", rejectMsg.Reason)
		rejectReason := rejectMsg.Reason
		if rejectReason == "" {
			rejectReason = "REJECTED"
		}
		c.Disconnect()
		c.onDisconnectHandler(rejectReason, nil)
	case "offer":
		offerMsg := webrtc.SessionDescription{}
		if err := unmarshalMessage(c, rawMessage, &offerMsg); err != nil {
			return err
		}
		if c.pc != nil && c.pc.SignalingState() == webrtc.SignalingStateHaveLocalOffer {
			c.createPeerConnection()
		}
		return c.setOffer(offerMsg)
	case "answer":
		answerMsg := webrtc.SessionDescription{}
		if err := unmarshalMessage(c, rawMessage, &answerMsg); err != nil {
			return err
		}
		return c.setAnswer(answerMsg)
	case "candidate":
		candidateMsg := candidateMessage{}
		if err := unmarshalMessage(c, rawMessage, &candidateMsg); err != nil {
			return err
		}
		if candidateMsg.ICECandidate != nil {
			c.trace("Received ICE candidate: %v", *candidateMsg.ICECandidate)
			c.addICECandidate(*candidateMsg.ICECandidate)
		}
	default:
		c.trace("invalid message type %s", message.Type)
		return errorInvalidMessageType
	}
	return nil
}
