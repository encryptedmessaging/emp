package network

import (
	zmq "github.com/alecthomas/gozmq"
	"fmt"
)

const (
	bufLen = 100
)

func publishLoop(frames chan Frame, log chan string, quit chan bool, pubSocket zmq.Socket) {
	defer pubSocket.Close()

	for {
		select {
		case <-quit:
			log <- "Ending Publishing Loop..."
			return
		case frm := <-frames:
			pubSocket.Send(frm.GetBytes(), 0)
		}
	}
}

func Publish(port uint16, log chan string) (chan Frame, chan bool) {
	pubSocket, err := context.NewSocket(zmq.PUB)
	if err != nil {
		log <- err
		return nil, nil
	}

	pubSocket.Bind(fmt.Sprintf("tcp://*:%d", port))

	quit := make(chan bool, 1)
	frames := make(chan Frame, bufLen)

	go publishLoop(frames, log, quit, pubSocket)
	return frames, quit
}
