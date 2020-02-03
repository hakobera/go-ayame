# go-ayame

go-ayame は [WebRTC Signaling Server Ayame](https://github.com/OpenAyame/ayame) の Go 言語用のクライアントライブラリです。

## 前提事項

go-ayame を利用するには Go 1.13 以上が必要です。

## 使い方

```go
import "github.com/hakobera/go-ayame/ayame"

signalingURL := "wss://ayame-lite.shiguredo.jp/signaling"
roomID := "your_room_id"

// ayame.Connect の作成
opts := ayame.DefaultOptions()
con := ayame.NewConnection(signalingURL, roomID, opts, false, false)

// PeerConnecition 接続時の処理
con.OnConnect(func() {
    fmt.Println("Connected")
})

// 動画、音声パケットデータ受信時の処理
con.OnTrackPacket(func(track *webrtc.Track, packet *rtp.Packet) {
    switch track.Kind() {
    case webrtc.RTPCodecTypeAudio:
        // Audio データを使って何かをする
    case webrtc.RTPCodecTypeVideo:
        // Video データを使って何かをする
    }
})

// Ayame サーバーへ接続
err := con.Connect()
if err != nil {
    log.Fatal("failed to connect Ayame", err)
}
```

## License

```
Copyright 2020 Kazuyuki Honda (hakobera)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
