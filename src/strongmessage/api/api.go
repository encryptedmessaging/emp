package api

import (
	"strongmessage/network"
	"strongmessage/objects"
	"time"
)

type ApiConfig struct {
	SendChan  chan network.Frame
	RecvChan  chan network.Frame
	RepRecv   chan network.Frame
	RepSend   chan network.Frame
	PeerChan  chan network.Peer
	LocalPeer *network.Peer
}

func Start(log chan string, config *ApiConfig, peers network.PeerList) {
	var frame *network.Frame

	// Create local version with which to connect to peers
	localVersion := new(objects.Version)

	localVersion.Version = uint32(objects.LOCAL_VERSION)
	localVersion.UserAgent = objects.LOCAL_USER
	localVersion.Timestamp = time.Now()
	localVersion.IpAddress = config.LocalPeer.IpAddress
	localVersion.Port = config.LocalPeer.Port
	localVersion.AdminPort = config.LocalPeer.AdminPort

	frame = network.NewFrame("version", localVersion.GetBytes(log))
	peers.SendAll(log, frame, config.RecvChan)

	for {
		select {
		case *frame = <-config.RecvChan:
			// Handle Received frames that do not require replies

		case *frame = <-config.RepRecv:
			// Handle requests that require replies to config.RepSend
		}
	}
}
