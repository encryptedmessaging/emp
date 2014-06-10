package objects

import (
  "fmt"
  "net"
  "time"
  zmq "github.com/alecthomas/gozmq"
)

type Peer struct {
  IpAddress net.IP
  Port uint16
  LastSeen time.Time
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
    socket.Connect(p.IpString())
  }
}
