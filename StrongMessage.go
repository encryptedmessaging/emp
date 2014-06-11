package main

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"strong-message/config"
	"strong-message/objects"
)

var LogChannel = make(chan string)
var MessageChannel = make(chan objects.Message)

func BootstrapNetwork(log chan string) {
	peers := config.LoadPeers(log)
	if peers == nil {
		log <- "Failed loading peers"
	} else {
		context, err := zmq.NewContext()
		if err != nil {
			log <- "Error creating ZMQ context"
			log <- err.Error()
		} else {
			for _, v := range peers {
				go v.Subscribe(LogChannel, MessageChannel, context)
			}
		}
	}
}

func StartPubServer(log chan string) {
	context, err := zmq.NewContext()
	if err != nil {
		log <- "Error creating ZMQ context"
		log <- err.Error()
	} else {
		socket, err := context.NewSocket(zmq.PUB)
		if err != nil {
			log <- "Error creating socket."
			log <- err.Error()
		}
		socket.Bind("tcp://127.0.0.1:5000")
		for {
			message := <-MessageChannel
			bytes := message.GetBytes(log)
			socket.Send(bytes, 0)
		}
	}
}

func BlockingLogger(channel chan string) {
	for {
		log_message := <-channel
		fmt.Println(log_message)
	}
}

func main() {
	go BootstrapNetwork(LogChannel)
	BlockingLogger(LogChannel)
}
