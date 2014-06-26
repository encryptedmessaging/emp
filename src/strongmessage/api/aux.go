package api

import (
	"fmt"
	"strongmessage/db"
	"strongmessage/network"
	"strongmessage/objects"
	"time"
)

func handleVersion(log chan string, config *ApiConfig, version *objects.Version, peers *network.PeerList) {
	newPeer := new(network.Peer)

	newPeer.IpAddress = version.IpAddress
	newPeer.Port = version.Port
	newPeer.AdminPort = version.AdminPort
	newPeer.LastSeen = time.Now()

	// Prepare Peer Request
	frame := new(network.Frame)
	frame.Magic = network.KNOWN_MAGIC
	frame.Type = "peer"

	for _, peer := range peers.Peers {
		frame.Payload = append(frame.Payload, peer.GetBytes()...)
		if peer.TcpString() == newPeer.TcpString() {
			peer.LastSeen = newPeer.LastSeen
			log <- fmt.Sprintf("Peer %s updated", peer.TcpString())
			config.PeerChan <- *newPeer
			newPeer.SendRequest(log, frame, config.RecvChan)
			return
		}
	}
	newPeer.Connect(log, config.Context)
	peers.Peers = append(peers.Peers, *newPeer)
	config.PeerChan <- *newPeer
	log <- fmt.Sprintf("New Peer Added: %s", newPeer.TcpString())

	// Send Peer Request
	newPeer.SendRequest(log, frame, config.RecvChan)
}

func handlePeer(log chan string, config *ApiConfig, newPeers []byte, peers *network.PeerList, localVersion *objects.Version, reqPeer *network.Peer) {

	// Prepare Version Request
	frame := new(network.Frame)
	frame.Magic = network.KNOWN_MAGIC
	frame.Type = "version"
	frame.Payload = localVersion.GetBytes(log)

	// Condense PeerList to Peer Hash Table
	peerHash := make(map[string]*network.Peer)

	for _, peer := range peers.Peers {
		peerHash[peer.TcpString()] = &peer
	}

	for i := 0; i <= len(newPeers)-28; i += 28 {
		p := new(network.Peer)
		err := p.FromBytes(newPeers[i : i+28])
		if err != nil {
			log <- fmt.Sprintf("Error unserializing peer... %s", err)
			continue
		}

		_, ok := peerHash[p.TcpString()]
		if !ok {
			peerHash[p.TcpString()] = p

			// Send Version Request
			err = p.Connect(log, config.Context)
			peers.Peers = append(peers.Peers, *p)
			if err != nil {
				p.SendRequest(log, frame, config.RecvChan)
			}
		}
	}

	// Send Object Request
	hashes := db.HashCache()
	frame.Type = "obj"
	frame.Payload = nil

	for _, hash := range hashes {
		frame.Payload = append(frame.Payload, []byte(hash)...)
	}
	reqPeer.SendRequest(log, frame, config.RecvChan)
}

func handleObj(log chan string, config *ApiConfig, objs []byte, reqPeer *network.Peer) {
	if reqPeer == nil {
		return
	}

	for i := 0; i <= len(objs)-48; i += 48 {
		if db.Contains(string(objs[i:i+48])) == db.NOTFOUND {
			frame := new(network.Frame)
			frame.Magic = network.KNOWN_MAGIC
			frame.Type = "getobj"
			copy(frame.Payload, objs[i:i+48])
			reqPeer.SendRequest(log, frame, config.RecvChan)
		}
	}

}
