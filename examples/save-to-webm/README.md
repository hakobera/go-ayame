# save-to-webm (go-ayame 版)

go-ayame と Pion、[Ayame Labo](https://ayame-labo.shiguredo.jp) を使って、WebRTC P2P 経由で受け取った Video と Audio データを WebM 形式の動画ファイルとして保存するサンプルコードです。

このアプリケーションは [pion/example-webrtc-applications](https://github.com/pion/example-webrtc-applications) 中の [save-to-web](https://github.com/pion/example-webrtc-applications/tree/master/save-to-webm) のシグナリングの部分を go-ayame を利用するように変更したものです。

## 使い方

### Ayame のオンラインサンプル「送信のみ(sendonly)」を開きます

Ayame のオンラインサンプル [送信のみ(sendonly)](https://openayame.github.io/ayame-web-sdk-samples/sendonly.html) をブラウザで開き、接続したいルームIDを入力して、接続ボタンを押します。

### save-to-web を実行します

上記で入力した RoomID をコマンドラインパラメータとして指定します。

```console
go run main.go -url wss://ayame-labo.shiguredo.jp/signaling -room-id <room-id>
```

PeerConnection 接続が完了すると、コンソールに `Connected` と表示され、ブラウザからの送信された動画と音声データが実行したフォルダ内の `test.webm` という名前のファイルに保存されます。

プログラムを終了するには、`Ctrl+C` を押します。プログラム終了が、動画プレイヤーで `test.webm` を再生して録画できていることを確認してください。

## 接続に失敗する場合の回避方法

Ayame のオンラインサンプルの sendonly のページを開いている host と、
save-to-webm コマンドを実行する host が同一だと接続に失敗する事例が報告されています。

https://github.com/hakobera/go-ayame/issues/10

その場合は、同じネットワーク上のちがう host から送信することで回避できます。
