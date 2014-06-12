package strongmessage

import (
	"strongmessage/network"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

func LoadPeers(log chan string) network.PeerList {
  var peer_list network.PeerList
  peers_filepath := "../peers.json"
  content, err := ioutil.ReadFile(peers_filepath)
  if err != nil {
    log <- "Errror opening " + peers_filepath
    log <- err.Error()
  } else {
    log <- "Loaded peers from: " + peers_filepath
    err = json.Unmarshal(content, &peer_list)
    if err != nil {
      log <- "Error parsing json"
      log <- err.Error()
    } else {
      msg := fmt.Sprintf("Loaded %d peers from config", len(peer_list.Peers))
      log <- msg
      return peer_list
    }
  }
  return peer_list
}

func BlockingLogger(channel chan string) {
  for {
    log_message := <-channel
    fmt.Println(log_message)
  }
}
