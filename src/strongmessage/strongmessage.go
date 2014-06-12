package strongmessage

import (
  "fmt"
  zmq "github.com/alecthomas/gozmq"
  "strongmessage/config"
  "strongmessage/network"
  "strongmessage/objects"
  "io/ioutil"
  "encoding/json"
)

func StartSubscriptions(log chan string, messageChannel chan objects.Message) {
  peers := LoadPeers(log)
  log <- fmt.Sprintf("%v", peers)
  context, err := zmq.NewContext()
  if err != nil {
    log <- "Error creating ZMQ context"
    log <- err.Error()
  } else {
    peerChannel = make(chan Peer)
    go func() {
      for {
        peer := <- peer_channel
      }
    }()
    peers.Subscribe(log, messageChannel, context)
  }
}

func StartPubServer(log chan string, frameChannel chan network.Frame) error {
  context, err := zmq.NewContext()
  if err != nil {
    log <- "Error creating ZMQ context"
    log <- err.Error()
    return err
  } else {
    socket, err := context.NewSocket(zmq.PUB)
    if err != nil {
      log <- "Error creating socket."
      log <- err.Error()
    }
    tcpString := fmt.Sprintf("tcp://%s:%d", config.DOMAIN, config.PORT)
    socket.Bind(tcpString)
    for {
      message := <- message_channel
      bytes := message.GetBytes(log)
      socket.Send(bytes, 0)
    }
    return nil
  }
}

func BlockingLogger(channel chan string) {
  for {
    log_message := <-channel
    fmt.Println(log_message)
  }
}

func LoadPeers(log chan string) network.PeerList {
  var peer_list network.PeerList
  peers_filepath := "./peers.json"
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
  //this should use error checking at some point
  return peer_list
}
