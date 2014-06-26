package main

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"net"
	"strongmessage"
	"strongmessage/api"
	"strongmessage/local/localapi"
	"strongmessage/network"
	"strongmessage/objects"
	"time"
)

const (
	bufLen = 10
)

func main() {
	log := make(chan string, 100)
	port := uint16(5000)
	repPort := uint16(5001)
	rpcPort := uint16(8080)

	// Start 0MQ Context
	context, err := zmq.NewContext()

	if err != nil {
		fmt.Println("Error creating ZMQ Context object.")
	}

	defer context.Close()

	// Create Channels
	recvChan := make(chan network.Frame, bufLen)
	sendChan := make(chan network.Frame, bufLen)

	repRecv := make(chan network.Frame)
	repSend := make(chan network.Frame)

	peerChan := make(chan network.Peer)

	// Start Subscription Service
	check, recvSocket := network.Subscription(log, recvChan, peerChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	defer recvSocket.Close()

	// Start Publish Service
	check, sendSocket := network.Publish(port, log, sendChan, context)
	if !check {
		fmt.Println("Could not start subscription service.")
		return
	}

	defer sendSocket.Close()

	// Start Reply Server
	check, repSocket := network.RepServer(repPort, log, repSend, repRecv, context)
	if !check {
		fmt.Println("Could not start reply server.")
		return
	}

	defer repSocket.Close()

	fmt.Println("Services started successfully!")

	fmt.Println("Connecting to peers...")
	//Load peers from peers.json
	peers, err := strongmessage.LoadPeers()
	if err != nil {
		fmt.Println("Could not load peers from peers.json")
		fmt.Println("%s", err.Error())
		// This is not fatal
	}
	peers.ConnectAll(log, context)
	defer peers.DisconnectAll()

	channels := new(api.ApiConfig)
	channels.SendChan = sendChan
	channels.RecvChan = recvChan
	channels.RepRecv = repRecv
	channels.RepSend = repSend
	channels.PeerChan = peerChan
	channels.Context = context
	channels.PubkeyRegister = make(chan []byte, bufLen)
	channels.MessageRegister = make(chan objects.Message, bufLen)
	channels.DBFile = "inventory.db"
	channels.LocalDB = "local.db"

	// Setup Local Peer
	channels.LocalPeer = new(network.Peer)
	channels.LocalPeer.IpAddress = net.ParseIP("10.50.10.109")
	channels.LocalPeer.Port = port
	channels.LocalPeer.AdminPort = repPort

	localVersion := new(objects.Version)

	localVersion.Version = uint32(objects.LOCAL_VERSION)
	localVersion.UserAgent = objects.LOCAL_USER
	localVersion.Timestamp = time.Now()
	localVersion.IpAddress = channels.LocalPeer.IpAddress
	localVersion.Port = channels.LocalPeer.Port
	localVersion.AdminPort = channels.LocalPeer.AdminPort

	channels.LocalVersion = localVersion

	err = localapi.Initialize(log, channels, rpcPort)

	if err != nil {
		fmt.Println("Could not start RPC Server: ", err.Error())
	}

	go api.Start(log, channels, &peers)

	fmt.Println("Connected... starting logger")
	strongmessage.BlockingLogger(log)
}
