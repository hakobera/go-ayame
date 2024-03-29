# SDL2

go-ayame と Pion、libvpx、[Ayame Labo](https://ayame-labo.shiguredo.jp) を使って、WebRTC P2P 経由で受け取った Video データを [SDL2](https://www.libsdl.org/) を使って表示するサンプルコードです。

## 使い方

### 依存ライブラリのインストール

`macOS + libvpx 1.8` と `Ubuntu 18.04.4 LTS + libvpx 1.7` の組み合わせでのみ動作確認しています。
それぞれの環境で以下のコマンドを実行して、依存ライブラリをインストールしてください。

#### macOS

```console
$ brew install sdl2 libvpx
```

[Homebrew](https://brew.sh) がセットアップされていない場合は、先にセットアップを済ませておいてください。

#### Ubuntu 18.04

```console
$ sudo apt install libsdl2-dev libvpx-dev
```

### Ayame のオンラインサンプル「送信のみ(sendonly)」を開きます

Ayame のオンラインサンプル [送信のみ(sendonly)](https://openayame.github.io/ayame-web-sdk-samples/sendonly.html) をブラウザで開き、接続したいルームIDを入力して、接続ボタンを押します。

### sdl2 を実行します

上記で入力した RoomID をコマンドラインパラメータとして指定します。

```console
go run . -url wss://ayame-labo.shiguredo.jp/signaling -room-id <room-id>
```

プログラムが開始されると、SDLのウィンドウが開き、PeerConnection 接続が完了すると、コンソールに `Connected` と表示されます。ブラウザからの送信された動画データは、初回キーフレームを受け取った後にSDLウィンドウ内に表示されます。

プログラムを終了するには、`Ctrl+C` を押します。

### VP9 を利用する

デフォルトでは VP8 形式の動画を受信するようになっていますが、このサンプルは VP9 形式の動画を受信することもできます。その場合、`-video-codec VP9` オプションを追加します。

```console
go run . -url wss://ayame-labo.shiguredo.jp/signaling -room-id <room-id> -video-codec VP9
```

### 詳細ログを出力する

詳細ログを出力する場合は、`-verbose` オプションを追加します。

```console
go run . -url wss://ayame-labo.shiguredo.jp/signaling -room-id <room-id> -verbose
```
