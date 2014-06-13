package network

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

func RepServer(port uint16, log chan string, sendChannel chan Frame, recvChannel chan Frame, context *zmq.Context) (bool, *zmq.Socket) {
	// Create REP Socket
	repSocket, err := context.NewSocket(zmq.REP)
	if err != nil {
		log <- "Error creating rep socket..."
		log <- err.Error()
		return false, nil
	}

	// Bind REP Socket
	err = repSocket.Bind(fmt.Sprintf("tcp://*:%d", port))
	if err != nil {
		log <- "Error binding pub socket..."
		log <- err.Error()
		repSocket.Close()
		return false, nil
	}

	// Start REP Loop
	go func() {
		for {
			var frame Frame
      data, err := repSocket.Recv(0)

      if err != nil {
        log <- "Error receiving from socket..."
        log <- err.Error()
      }

      frame, err = FrameFromBytes(data)
      if err != nil {
        log <- "Received invalid frame..."
        log <- err.Error()
				repSocket.Send(nil, 0)
				continue
      }

			// Should block until mux is ready...
      recvChannel <- frame

			// Should block until reply is ready...
			frame = <-sendChannel

			err = repSocket.Send(frame.GetBytes(), 0)
			if err != nil {
				log <- "Error sending frame..."
				log <- err.Error()
			}
		}
	}()

	return true, repSocket
}
