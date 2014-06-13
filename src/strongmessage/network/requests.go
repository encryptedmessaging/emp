package network

import (
	zmq "github.com/alecthomas/gozmq"
)

func ReqClient(log chan string, sendChannel chan Frame, recvChannel chan Frame, peerChannel chan Peer, context *zmq.Context) (bool, *zmq.Socket) {
	socket, err := context.NewSocket(zmq.REQ)
	if err != nil {
		log <- "error creating socket"
		log <- err.Error()
		return false, nil
	}

	// Peer Channel Request Loop
	go func() {
		var err error
		for {
			peer := <-peerChannel
			err = socket.Connect(peer.TcpString())
			if err != nil {
				log <- "Error subscribing to peer..."
				log <- err.Error()
			}

		}
	}()

	// Frame Channel Request/Receive Loop
	go func() {
		for {
			var frame Frame

			frame = <-sendChannel

			err = socket.Send(frame.GetBytes(), 0)
      if err != nil {
        log <- "Error sending frame..."
        log <- err.Error()
      }

			data, err := socket.Recv(0)

			if err != nil {
				log <- "Error receiving from socket..."
				log <- err.Error()
			}

			frame, err = FrameFromBytes(data)
			if err != nil {
				log <- "Received invalid frame..."
				log <- err.Error()
			}
			recvChannel <- frame
		}
	}()

	return true, socket
}
