package ayame

import "github.com/pion/webrtc/v2"

type message struct {
	Type string `json:"type"`
}

type pingMessage struct {
	Type string `json:"type"`
}

type pongMessage struct {
	Type string `json:"type"`
}

type registerMessage struct {
	Type          string       `json:"type"`
	RoomID        string       `json:"roomId"`
	ClientID      string       `json:"clientId"`
	AuthnMetadata *interface{} `json:"authnMetadata"`
	SignalingKey  *string      `json:"signalingKey"`

	// Ayame クライアント情報が詰まっている
	AyameClient *string `json:"ayameClient"`
	Environment *string `json:"environment"`
}

type byeMessage struct {
	Type string `json:"type"`
}

type acceptMessage struct {
	Type          string       `json:"type"`
	ConnectionID  string       `json:"connectionId"`
	AuthzMetadata *interface{} `json:"authzMetadata,omitempty"`
	IceServers    *[]iceServer `json:"iceServers,omitempty"`
	IsExistClient bool         `json:"isExistClient"`
}

type rejectMessage struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type candidateMessage struct {
	Type         string                   `json:"type"`
	ICECandidate *webrtc.ICECandidateInit `json:"ice,omitempty"`
}

type iceServer struct {
	Urls       []string `json:"urls"`
	UserName   *string  `json:"username,omitempty"`
	Credential *string  `json:"credential,omitempty"`
}
