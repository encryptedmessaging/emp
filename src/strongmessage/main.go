package main

import (
	"strongmessage/network"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

func BlockingLogger(channel chan string) {
  for {
    log_message := <-channel
    fmt.Println(log_message)
  }
}

func main() {
	log := make(chan string, 100)
	port := uint16(4444)

	context, err := zmq.NewContext()
	defer context.Close()

	if err != nil {
		fmt.Println("Error creating ZMQ Context object.")
	}

	recvChan := make(chan network.Frame)
	sendChan := make(chan network.Frame)

	peerChan := make(chan network.Peer)

	check := network.Subscription(log, recvChan, peerChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}
	check = network.Publish(port, log, sendChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	fmt.Println("Services started successfully!")
	BlockingLogger(log)
}
