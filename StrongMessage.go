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

func route(from, to1, to2 chan network.Peer) {
	var p network.Peer
	for {
		p = <-from
		to1 <- p
		to2 <- p
	}
}

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

	reqSend := make(chan network.Frame)

	peerChan := make(chan network.Peer)
	subPeer := make(chan network.Peer)
	reqPeer := make (chan network.Peer)

	go route(peerChan, subPeer, reqPeer)

	// Start Subscription Service
	check, recvSocket := network.Subscription(log, recvChan, subPeer, context)
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

	// Start Request Service
	check, reqSocket := network.ReqClient(log, reqSend, recvChan, reqPeer, context)
	if !check {
		fmt.Println("Could not start request service.")
    return
	}

	defer reqSocket.Close()

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
