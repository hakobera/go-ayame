# save-to-webm (go-ayame 版)

go-ayame と Pion を使って、[Ayame Lite](https://ayame-lite.shiguredo.jp/beta) を使って、WebRTC P2P で DataChannel を使用したテキストデータのやりとりをするサンプルコードです。

## 使い方

### Ayame のオンラインサンプル「RTCDataChannel」を開きます

Ayame のオンラインサンプル [RTCDataChannel)](https://openayame.github.io/ayame-web-sdk-samples/datachannel.html) をブラウザで開き、接続したいルームIDを入力して、接続ボタンを押します。

### datachannel を実行します

上記で入力した RoomID をコマンドラインパラメータとして指定します。

```console
go run main.go -url wss://ayame-lite.shiguredo.jp/signaling -room-id <room-id>
```

PeerConnection 接続が完了すると、コンソールに `Connected` と表示されます。

go-ayame では、P2P の相手が作成した DataChannel および自身が作成した DataChannel の両方を扱うことができます。このサンプルでは、これを示すために、2つの DataChannel が作成されます。

- label=`dataChannel`

Ayame のオンラインサンプルがブラウザ上で作成される DataChannel です。サンプル上の「送信するメッセージ」テキストボックスに任意の文字を入力して、「送信」ボタンを押すと、ブラウザからサンプルプログラムに対して、テキストメッセージが送信されます。サンプルでは、送られてきたメッセージに `[echo]` を文頭に追加して、そのまま送り返しています。結果は、ブラウザ上の「受信したメッセージ」テキストエリアで確認することができます。

- label=`goAyameExample`

サンプルプログラムで `ayame.Connect.AddDataChannel()` メソッドを利用して作成される DataChannel です。こちらは、5秒おきに `[push] Hello DataChannel` というテキストを送信し続けます。結果は、ブラウザ上の「受信したメッセージ」テキストエリアで確認することができます。。

プログラムを終了するには、`Ctrl+C` を押します。
