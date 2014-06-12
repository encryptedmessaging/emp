package strongmessage

import (
  "fmt"
  zmq "github.com/alecthomas/gozmq"
  "strongmessage/config"
  "strongmessage/objects"
)

func BootstrapNetwork(log chan string, message_channel chan objects.Message) {
  peers := config.LoadPeers(log)
  log <- fmt.Sprintf("%v", peers)
  if peers == nil {
    log <- "Failed to load peers.json"
  } else {
    context, err := zmq.NewContext()
    if err != nil {
      log <- "Error creating ZMQ context"
      log <- err.Error()
    } else {
      for _, v := range peers {
        go v.Subscribe(log, message_channel, context)
      }
    }
  }
}

func StartPubServer(log chan string, message_channel chan objects.Message) error {
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
