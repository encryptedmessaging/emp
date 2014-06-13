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
	strongmessage.BlockingLogger(log)
}
