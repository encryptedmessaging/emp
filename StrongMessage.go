package main

import (
  "strongmessage"
  "strongmessage/objects"
)

var LogChannel = make(chan string)
var FrameChannel = make(chan network.Frame)



func main() {
	go strongmessage.StartSubscriptions(LogChannel, FrameChannel)
	go strongmessage.StartPubServer(LogChannel, FrameChannel)
  strongmessage.BlockingLogger(LogChannel)
}
