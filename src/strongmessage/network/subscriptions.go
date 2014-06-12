package network


func Subscription(log chan string, frameChannel chan Frame, peerChannel chan Peer) bool {
  socket, err := context.NewSocket(zmq.SUB)
  if err != nil {
    log <- "error creating socket"
    log <- err.Error()
    return false
  }
  go func() {
    for {
      peer := <- peerChannel
      socket.Connect(peer.TcpString())
    }
  }
  return true
}
