package ayame

import "github.com/pion/webrtc/v3"

// ConnectionOptions は Ayame 接続オプションです。
type ConnectionOptions struct {
	// Audio の設定
	Audio ConnectionAudioOption

	// Video の設定
	Video ConnectionVideoOption

	// クライアントID
	ClientID string

	// Ayame から ICEServer 情報が返って来なかった場合に使われる ICEServer の情報
	ICEServers []webrtc.ICEServer

	// 認証が必要なルームへの接続時に必要なシグナリングキー
	SignalingKey string

	// TrickleICE を利用するかどうかのフラグ
	UseTrickeICE bool
}

// ConnectionVideoOption は Video に関するオプションです。
type ConnectionVideoOption struct {
	// 送受信方向。現在、'recvonly' のみサポート
	Direction webrtc.RTPTransceiverDirection

	// 有効かどうかのフラグ
	Enabled bool

	// 対応しているコーデックの設定
	Codecs []*webrtc.RTPCodecParameters
}

// ConnectionAudioOption は Audio に関数するオプションです。
type ConnectionAudioOption struct {
	// 送受信方向。現在、'recvonly' のみサポート
	Direction webrtc.RTPTransceiverDirection

	// 有効かどうかのフラグ
	Enabled bool

	// 対応しているコーデックの設定
	Codecs []*webrtc.RTPCodecParameters
}
