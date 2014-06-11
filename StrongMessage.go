package main

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"strong-message/config"
	"strong-message/objects"
)

var LogChannel = make(chan string)
var MessageChannel = make(chan objects.Message)

func BlockingLogger(channel chan string) {
	for {
		log_message := <-channel
		fmt.Println(log_message)
	}
}

func main() {
	go strongMessage.BootstrapNetwork(LogChannel)
	BlockingLogger(LogChannel)
}
