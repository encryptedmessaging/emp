package config

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "time"
)

type PeerList struct {
  Peers []Peer `json:"peers"`
}

type Peer struct {
  IpAddress string `json:"ip_address"`
  LastSeen time.Time `json:"last_seen"`
}

func LoadPeers(log chan string) {
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
    }
  }
}
