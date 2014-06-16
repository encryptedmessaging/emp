package network

import (
	zmq "github.com/alecthomas/gozmq"
	"fmt"
)

func Subscription(log chan string, frameChannel chan Frame, peerChannel chan Peer, context *zmq.Context) (bool, *zmq.Socket) {
	socket, err := context.NewSocket(zmq.SUB)
	if err != nil {
		log <- "error creating socket"
		log <- err.Error()
		return false, nil
	}

	// Peer Channel Subscription Loop
	socket.SetSubscribe("")
	go func() {
		var err error
		for {
			peer := <-peerChannel
			err = socket.Connect(peer.TcpString())
			if err != nil {
				log <- fmt.Sprintf("Error subscribing to peer...%s; %s", peer.TcpString(), err.Error())
			}

		}
	}()

	// Frame Channel Receive Loop
	go func() {
		for {
			var frame Frame
			data, err := socket.Recv(0)

			if err != nil {
				log <- fmt.Sprintf("Error receiving from subscription socket... %s", err.Error())
				continue
			}

			frame, err = FrameFromBytes(data)
			if err != nil {
				log <- fmt.Sprintf("Received invalid frame... %s", err.Error())
			}
			frameChannel <- frame
		}
	}()

	return true, socket
}
