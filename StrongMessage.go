package main

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"strongmessage"
	"strongmessage/network"
	"strongmessage/api"
	"os"
)

const (
	bufLen = 10
)

func main() {
	log := make(chan string, 100)
	port := uint16(4444)
	repPort := uint16(4445)

	// Start 0MQ Context
	context, err := zmq.NewContext()

	if err != nil {
		fmt.Println("Error creating ZMQ Context object.")
	}

	defer context.Close()

	// Create Channels
	recvChan := make(chan network.Frame, bufLen)
	sendChan := make(chan network.Frame, bufLen)

	repRecv := make(chan network.Frame)
	repSend := make(chan network.Frame)

	peerChan := make(chan network.Peer)

	// Start Subscription Service
	check, recvSocket := network.Subscription(log, recvChan, peerChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	defer recvSocket.Close()

	// Start Publish Service
	check, sendSocket := network.Publish(port, log, sendChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	defer sendSocket.Close()

	// Start Reply Server
	check, repSocket := network.RepServer(repPort, log, repRecv, repSend, context)
	if !check {
		fmt.Println("Could not start reply server.")
		return
	}

	defer repSocket.Close()

	fmt.Println("Services started successfully!")

	fmt.Println("Connecting to peers...")
	//Load peers from peers.json
	peers, err := strongmessage.LoadPeers()
	if err != nil {
		fmt.Println("Could not load peers from peers.json")
		fmt.Println("%s", err.Error())
		// This is not fatal
	}
	peers.ConnectAll(log, context)
	defer peers.DisconnectAll()

	// Setup Signals
	quit := make(chan os.Signal, 1)

	channels := new(api.ApiConfig)
	channels.SendChan = sendChan
	channels.RecvChan = recvChan
	channels.RepRecv = repRecv
	channels.RepSend = repSend
	channels.PeerChan = peerChan

	go api.Start(log, channels, peers)

	fmt.Println("Connected... starting logger")
	strongmessage.BlockingLogger(log, quit)
}
