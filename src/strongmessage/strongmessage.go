package strongmessage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strongmessage/network"
	"os"
)

func LoadPeers() (network.PeerList, error) {
	var peerList network.PeerList
	peersPath := "./peers.json"
	content, err := ioutil.ReadFile(peersPath)
	if err != nil {
		return peerList, err
	}
	err = json.Unmarshal(content, &peerList)
	if err != nil {
		return peerList, err
	}
	return peerList, nil
}

func BlockingLogger(channel chan string, quit chan os.Signal) {
	for {
		select {
		case log_message := <-channel:
			fmt.Println(log_message)
		case <-quit:
			fmt.Println("Exiting...")
			return
		}
	}
}
