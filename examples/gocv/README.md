# GoCV

go-ayame と Pion、libvpx、[Ayame Lite](https://ayame-lite.shiguredo.jp/beta) を使って、WebRTC P2P 経由で受け取った Video データを使ってモーション検知をし、検知した場合に DataChannel を利用して送信側にデータを送信するサンプルコードです。

## 使い方

### 依存ライブラリのインストール

`macOS + libvpx 1.8` と `Ubuntu 18.04.4 LTS + libvpx 1.7` の組み合わせでのみ動作確認しています。
それぞれの環境で以下のコマンドを実行して、依存ライブラリをインストールしてください。

#### macOS

```console
$ brew install libvpx
```

[Homebrew](https://brew.sh) がセットアップされていない場合は、先にセットアップを済ませておいてください。

#### Ubuntu 18.04

```console
$ sudo apt install libvpx-dev libcanberra-gtk-module
```

### GoCV

それぞれの環境でのインストール方法については以下の URL を参照してください。

* macOS: https://gocv.io/getting-started/macos/
* Ubuntu: https://gocv.io/getting-started/linux/
