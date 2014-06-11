package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strong-message/objects"
)

type PeerList struct {
	Peers []objects.Peer `json:"peers"`
}

func LoadPeers(log chan string) []objects.Peer {
	peers_filepath := "./peers.json"
	content, err := ioutil.ReadFile(peers_filepath)
	if err != nil {
		log <- "Errror opening " + peers_filepath
		log <- err.Error()
	} else {
		log <- "Loaded peers from: " + peers_filepath
		var peer_list PeerList
		err = json.Unmarshal(content, &peer_list)
		if err != nil {
			log <- "Error parsing json"
			log <- err.Error()
		} else {
			msg := fmt.Sprintf("Loaded %d peers from config", len(peer_list.Peers))
			log <- msg
			return peer_list.Peers
		}
	}
	return nil
}
