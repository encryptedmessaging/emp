package objects

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"net"
	"time"
)

const DEBUG = true

type Peer struct {
	IpAddress net.IP    `json:"ip_address"`
	Port      uint16    `json:"port"`
	LastSeen  time.Time `json:"last_seen"`
}

func (p *Peer) IpString() string {
	return fmt.Sprintf("tcp://%s:%d", p.IpAddress.String(), p.Port)
}

func (p *Peer) Subscribe(log chan string, messageChannel chan Message, context *zmq.Context) {
	socket, err := context.NewSocket(zmq.SUB)
	if err != nil {
		log <- "Error creating socket"
		log <- err.Error()
	} else {
    log <- fmt.Sprintf("Attempting subscription: %s:%d", p.IpAddress, p.Port)
		socket.Connect(p.IpString())
    for {
      log <- fmt.Sprintf("Connected: %s:%d", p.IpAddress, p.Port)
      msg, err := socket.Recv(0)
      if err != nil {
        log <- "Socket Error:"
        log <- err.Error()
      } else {
        message, err := MessageFromBytes(log, msg)
        if err != nil {
          log <- "Decoding error:"
          log <- err.Error()
        } else {
          if DEBUG == true {
            log <- "Got message:"
            log <- fmt.Sprintf("%v", message)
          }
          messageChannel <- message
        }
      }
    }
	}
}
