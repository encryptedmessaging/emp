package main

import (
  "strongmessage/network"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"net"
	"time"
	"testing"
)

func TestServices(t *testing.T) {
	log := make(chan string, 100)
  port := uint16(4444)

  context, err := zmq.NewContext()
  
  if err != nil {
    fmt.Println("Error creating ZMQ Context object.")
		t.FailNow()
  }

	defer context.Close()

  recvChan := make(chan network.Frame, bufLen)
  sendChan := make(chan network.Frame, bufLen)

  peerChan := make(chan network.Peer)

  check, recvSocket := network.Subscription(log, recvChan, peerChan, context)
  if !check {
    fmt.Println("Could not start subscription service.")
		t.FailNow()
  }

	defer recvSocket.Close()

  check, sendSocket := network.Publish(port, log, sendChan, context)
  if !check {
    fmt.Println("Could not start subscription service.")
    t.FailNow()
	}

	defer sendSocket.Close()

	p := new(network.Peer)
  p.IpAddress = net.ParseIP("127.0.0.1")
  p.Port = uint16(4444)
  p.LastSeen = time.Now()

  peerChan <- *p
	time.Sleep(time.Millisecond)

  f := new(network.Frame)
  f.Magic = [4]byte{'a', 'b', 'c', 'd'}
  f.Type = [8]byte{'v', 'e', 'r', 's', 'i', 'o', 'n', '?'}
  f.Payload = []byte("Hello World!")

  sendChan <- *f

  f2 := <-recvChan

	if string(f.GetBytes()) != string(f2.GetBytes()) {
		fmt.Println("Received Message differs from Sent Message")
		t.Fail()
	} else {
		fmt.Println("Services Started Successfully!")
	}

}
