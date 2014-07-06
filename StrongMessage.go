package main

import (
	"fmt"
	"strongmessage"
	"strongmessage/api"
	"strongmessage/local/localapi"
	"quibit"
	"strongmessage/objects"
	"os"
	"os/signal"
)

const (
	bufLen = 10
)

func main() {
	config := new(api.ApiConfig)

	// Initialize All Config Options

	// Network Channels
	config.RecvQueue = make(chan quibit.Frame, bufLen)
	config.SendQueue = make(chan quibit.Frame, bufLen)
	config.PeerQueue = make(chan quibit.Peer)

	// Local Logic
	config.DbFile  = "inventory.db"
	config.LocalDB = "local.db"

	config.LocalVersion.Port = uint16(4444)
	config.RPCPort           = uint16(8080)

	// Local Registers
	config.PubkeyRegister  = make(chan objects.Hash, bufLen)
	config.MessageRegister = make(chan objects.Message, bufLen)
	config.PurgeRegister   = make(chan [16]byte, bufLen)

	// Administration
	config.Log  = make(chan string, bufLen)
	config.Quit = make(chan os.Signal, 1)

	// Start Network Services
	err := quibit.Initialize(config.Log, config.RecvQueue, config.SendQueue, config.PeerQueue, config.LocalVersion.Port)
	defer quibit.Cleanup()
	if err != nil {
		config.Log <- fmt.Sprintf("Error initializing network: %s", err)
		return
	}

	// Start Signal Handler
	signal.Notify(config.Quit, os.Interrupt, os.Kill)

	// Start API
	go api.Start(config)

	go localapi.Initialize(config)

	// Start Logger
	fmt.Println("Starting logger...")
	strongmessage.BlockingLogger(config.Log)

	// Give some time for cleanup...
	fmt.Println("Cleaning up...")
	localapi.Cleanup()
}
