package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/pion/webrtc/v3"
)

func main() {
	signalingURL := flag.String("url", "wss://ayame-labo.shiguredo.jp/signaling", "Specify Ayame service address")
	roomID := flag.String("room-id", "", "specify room ID")
	signalingKey := flag.String("signaling-key", "", "specify signaling key")
	verbose := flag.Bool("verbose", false, "enable verbose log")

	flag.Parse()
	log.Printf("args: url=%s, roomID=%s, signalingKey=%s", *signalingURL, *roomID, *signalingKey)

	opts := ayame.DefaultOptions()
	opts.SignalingKey = *signalingKey

	con := ayame.NewConnection(*signalingURL, *roomID, opts, *verbose, false)
	defer con.Disconnect()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dc *webrtc.DataChannel
	con.OnOpen(func(metadata *interface{}) {
		fmt.Println("Open")
		var err error
		dc, err = con.CreateDataChannel("go-ayame-example", nil)
		if err != nil {
			log.Printf("CreateDataChannel error: %v", err)
			return
		}
		log.Printf("CreateDataChannel: label=%s", dc.Label())
		dc.OnMessage(onMessage(dc))
	})

	con.OnConnect(func() {
		fmt.Println("Connected")
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			for {
				select {
				case <-ticker.C:
					if dc != nil {
						dc.SendText("[push] Hello DataChannel")
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	})

	con.OnDataChannel(func(c *webrtc.DataChannel) {
		log.Printf("OnDataChannel: label=%s", c.Label())
		if dc == nil {
			dc = c
			dc.OnMessage(onMessage(dc))
		}
	})

	err := con.Connect()
	if err != nil {
		log.Fatal("failed to connect Ayame", err)
	}

	fmt.Println("Press Ctrl+C to stop process")
	waitInterrupt()
	cancel()
	con.Disconnect()

	log.Printf("Done")
}

func waitInterrupt() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		<-interruptChannel
		log.Printf("Interrupt")
		endWaiter.Done()
	}()
	endWaiter.Wait()
}

func onMessage(dc *webrtc.DataChannel) func(webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			fmt.Printf("OnData[%s]: data=%s\n", dc.Label(), (msg.Data))
			dc.SendText("[echo] " + string(msg.Data))
		}
	}
}
