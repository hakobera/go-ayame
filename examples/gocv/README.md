# GoCV

go-ayame と Pion、libvpx、[Ayame Lite](https://ayame-lite.shiguredo.jp/beta) を使って、WebRTC P2P 経由で受け取った Video データを使ってモーション検知をし、検知した場合に DataChannel を利用して送信側にデータを送信するサンプルコードです。

実際に動いている様子は以下のリンク先の Tweet の動画で確認できます。

> Raspberry Pi Zero W + WebRTC Native Client momo で配信した動画を、go-ayame を使って受信して、GoCV (OpenCV) でモーション検知して、検知結果を DataChannel で戻して serial 経由で Arduino に指令を出して LED を点灯、消灯している図

https://twitter.com/hakobera/status/1244279413329416192

[![image](https://pbs.twimg.com/ext_tw_video_thumb/1244279387446341632/pu/img/_NJsxhk3Q1O1HegR?format=jpg&name=small)](https://twitter.com/hakobera/status/1244279413329416192)

## 依存ライブラリのインストール

`macOS + libvpx 1.8` と `Ubuntu 18.04.4 LTS + libvpx 1.7` の組み合わせでのみ動作確認しています。
それぞれの環境で以下のコマンドを実行して、依存ライブラリをインストールしてください。

### macOS

```console
$ brew install libvpx
```

[Homebrew](https://brew.sh) がセットアップされていない場合は、先にセットアップを済ませておいてください。

GoCV のインストール方法は[公式ドキュメント](https://gocv.io/getting-started/macos)を参照してください。

### Ubuntu 18.04

```console
$ sudo apt install libvpx-dev libcanberra-gtk-module
```

GoCV のインストール方法は[公式ドキュメント](https://gocv.io/getting-started/linux)を参照してください。

## 使い方

[WebRTC Native Client Momo](https://github.com/shiguredo/momo) と組み合わせて使うことを想定しています。

### Arduino

Arduino IDE などを使って、以下のスケッチを Arduino に書き込んでおきます。
ここでは 12 番のピンに LED をつなげている前提なので、必要に応じて `#define LED_1 12` の部分は書き換えてください。配線図については省力します。

```arduino
#define LED_1 12

void setup() {
  pinMode(LED_1, OUTPUT);
  Serial.begin(9600);
}

void loop() {
  int input;

  input = Serial.read();
  if (input != -1) {
    switch (input) {
      case 'o':
        Serial.print("LED1 ON\n");
        digitalWrite(LED_1, HIGH);
        break;

      case 'p':
        Serial.print("LED1 OFF\n");
        digitalWrite(LED_1, LOW);
        break;
    }
  }
}
```

### momo

momo を以下のオプション付きで起動します。
 
```console
$ momo --no-audio ayame wss://ayame-lite.shiguredo.jp/signaling <room-id> --serial /dev/<your-arduino-serial-device-name>,9600
```

### gocv

momo の起動オプションでしていした RoomID をコマンドラインパラメータとして指定します。

```console
go run . -url wss://ayame-lite.shiguredo.jp/signaling -room-id <room-id>
```
