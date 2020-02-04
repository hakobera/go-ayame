package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/hakobera/go-ayame/ayame"
	"github.com/pion/webrtc/v2"
)

func main() {
	signalingURL := flag.String("url", "wss://ayame-lite.shiguredo.jp/signaling", "Specify Ayame service address")
	roomID := flag.String("room-id", "", "specify room ID")
	signalingKey := flag.String("signaling-key", "", "specify signaling key")
	verbose := flag.Bool("verbose", false, "enable verbose log")

	flag.Parse()
	log.Printf("args: url=%s, roomID=%s, signalingKey=%s", *signalingURL, *roomID, *signalingKey)

	opts := ayame.DefaultOptions()
	opts.SignalingKey = *signalingKey
	con := ayame.NewConnection(*signalingURL, *roomID, opts, *verbose, false)

	con.OnConnect(func() {
		fmt.Println("Connected")
		dc, err := con.AddDataChannel("go-ayame-example", nil)
		if err != nil {
			log.Printf("AddDataChannel error: %v", err)
			return
		}
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			for range ticker.C {
				dc.SendText("Hello DataChannel")
			}
		}()
	})

	con.OnData(func(dc *webrtc.DataChannel, msg *webrtc.DataChannelMessage) {
		if msg.IsString {
			fmt.Printf("OnData: isString=true, data=%s\n", string(msg.Data))
			dc.SendText("[echo] " + string(msg.Data))
		}
	})

	err := con.Connect()
	if err != nil {
		log.Fatal("failed to connect Ayame", err)
	}

	fmt.Println("Press Ctrl+C to stop process")
	waitInterrupt()
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
