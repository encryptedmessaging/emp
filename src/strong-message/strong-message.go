package strongMessage

import (
	zmq "github.com/alecthomas/gozmq"
	"strong-message/objects"
)

func BoostrapNetwork(log_channel chan string, message_channel chan objects.Message) error {
	peers := loadPeers(log_channel)
	if peers == nil {
		log_channel <- "Failed to load peers.json"
	} else {
		context, err := zmq.NewContext()
		if err != nil {
			log_channel <- "Error creating ZMQ context"
			log_channel <- err.Error()
			return err
		} else {
			for _, v := range peers {
				go v.Subscribe(log_channel, message_channel, context)
			}
		}
		return nil
	}
}
