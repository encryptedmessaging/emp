package network

import (
  zmq "github.com/alecthomas/gozmq"
  "fmt"
)

const DEBUG = true

type PeerList struct {
  Peers []Peer `json:"peers"`
}

func (p *PeerList) Subscribe(log chan string, frameChannel chan Frame, context *zmq.Context) {
  socket, err := context.NewSocket(zmq.SUB)
  if err != nil {
    log <- "Error creating socket"
    log <- err.Error()
  } else {
    for _, v := range p.Peers {
      go func() {
        log <- fmt.Sprintf("Attempting subscription: %s:%d", v.IpAddress, v.Port)
        socket.Connect(v.TcpString())
        for {
          log <- fmt.Sprintf("Connected: %s:%d", v.IpAddress, v.Port)
          data, err := socket.Recv(0)
          if err != nil {
            log <- "Socket Error:"
            log <- err.Error()
          } else {
            frame, err := FrameFromBytes(data)
            if err != nil {
              log <- "Decoding error:"
              log <- err.Error()
            } else {
              if DEBUG == true {
                log <- "Got message:"
                log <- fmt.Sprintf("%v", frame)
              }
              frameChannel <- frame
            }
          }
        }

      }()
    }
  }
}

func LoadPeers(log chan string) PeerList {
  var peer_list PeerList
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
