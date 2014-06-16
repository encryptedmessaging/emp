package network

import (
	zmq "github.com/alecthomas/gozmq"
)

type PeerList struct {
	Peers []Peer `json:"peers"`
}

func (plist *PeerList) ConnectAll(log chan string, context *zmq.Context) {
	for _, peer := range plist.Peers {
		peer.Connect(log, context)
	}
}

func (plist *PeerList) SendAll(log chan string, frame *Frame, recvChannel chan Frame) {
	for _, peer := range plist.Peers {
		peer.SendRequest(log, frame, recvChannel)
	}
}

func (plist *PeerList) DisconnectAll() {
	for _, peer := range plist.Peers {
		peer.Disconnect()
	}
}
