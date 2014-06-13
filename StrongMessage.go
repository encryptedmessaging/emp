package main

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"strongmessage"
	"strongmessage/network"
)

const (
	bufLen = 10
)

func main() {
	log := make(chan string, 100)
	port := uint16(4444)

	context, err := zmq.NewContext()

	if err != nil {
		fmt.Println("Error creating ZMQ Context object.")
	}

	defer context.Close()

	recvChan := make(chan network.Frame, bufLen)
	sendChan := make(chan network.Frame, bufLen)

	peerChan := make(chan network.Peer)

	check, recvSocket := network.Subscription(log, recvChan, peerChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	defer recvSocket.Close()

	check, sendSocket := network.Publish(port, log, sendChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	defer sendSocket.Close()

	fmt.Println("Services started successfully!")

	fmt.Println("Connecting to peers...")
	//Load peers from peers.json
	peers, err := strongmessage.LoadPeers()
	if err != nil {
		fmt.Println("Could not load peers from peers.json")
		fmt.Println("%s", err.Error())
		// This is not fatal
	}
	for _, p := range peers.Peers {
		peerChan <- p
	}

	strongmessage.BlockingLogger(log)
}
