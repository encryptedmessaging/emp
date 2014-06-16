package api

import (
	"strongmessage/network"
)

type ApiConfig struct {
	SendChan chan network.Frame
	RecvChan chan network.Frame
	RepRecv chan network.Frame
	RepSend chan network.Frame
	PeerChan chan network.Peer
}

func Start(log chan string, config *ApiConfig, peers network.PeerList) {


	for {
		select {
			case <-config.RecvChan:
				// Handle Received frames that do not require replies
			case <-config.RepRecv:
				// Handle requests that require replies to config.RepSend
		}
	}
}