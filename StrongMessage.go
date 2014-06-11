package main

import (
  "strongmessage"
  "strongmessage/objects"
)

var LogChannel = make(chan string)
var MessageChannel = make(chan objects.Frame)



func main() {
	go strongmessage.BootstrapNetwork(LogChannel, MessageChannel)
	strongmessage.BlockingLogger(LogChannel)
}
