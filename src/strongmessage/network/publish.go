package network

import (
	zmq "github.com/alecthomas/gozmq"
)

func Publish(port uint16, log chan string, sendChannel chan Frame, context *zmq.Context) bool {
	// Create PUB Socket
	pubSocket, err := context.NewSocket(zmq.PUB)
	if err != nil {
		log <- "Error creating pub socket..."
		log <- err.Error()
		return false
	}

	// Bind PUB Socket
	err = pubSocket.Bind(fmt.Sprintf("tcp://*:%d", port))
	if err != nil {
		log <- "Error binding pub socket..."
		log <- err.Error()
		return false
	}

	// Start PUB Loop
	go func() {
		for {
			frame := <-sendChannel
			err := pubSocket.Send(frame.GetBytes(), 0)
			if err != nil {
				log <- "Error sending frame..."
				log <- err.Error()
			}
		}
	}()

	return true
}
