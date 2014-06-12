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

