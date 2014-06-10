package main

import (
  "fmt"
)

var LogChannel = make(chan string)

func BlockingLogger(channel chan string) {
  for {
    log_message := <- channel
    fmt.Println(log_message)
  }
}

func main() {
  BlockingLogger(LogChannel)
}
