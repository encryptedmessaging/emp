package main

import (
  "fmt"
  "strong-message/config"
)

var LogChannel = make(chan string)

func BootstrapNetwork (log chan string) {
  config.LoadPeers(log)
}

func BlockingLogger(channel chan string) {
  for {
    log_message := <- channel
    fmt.Println(log_message)
  }
}

func main() {
  go BootstrapNetwork(LogChannel)
  BlockingLogger(LogChannel)
}
