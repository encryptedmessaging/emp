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

	recvChannel, recvQuit := network.Publish(port, log, context)

	BlockingLogger(log)
}
