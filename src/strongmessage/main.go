package main

import (
	"strongmessage"
	"strongmessage/network"
)

func main() {
	log := make(chan string, 100)
	port := 4444

	context, err := zmq.NewContext()
	defer context.Close()

	if err != nil {
		fmt.Println("Error creating ZMQ Context object.")
	}

	recvChan := make(chan network.Frame)
	sendChan := make(chan network.Frame)

	peerChan := make(chan network.Peer)

	check := Subscription(log, recvChan, peerChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}
	check = Publish(port, log, sendChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	fmt.Println("Services started successfully!")
	BlockingLogger(log)
}
