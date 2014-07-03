package main

import (
	"fmt"
	"strongmessage"
	"strongmessage/api"
	"quibit"
	"time"
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
	config.DbFile = "inventory.db"

	config.LocalVersion.Port = uint16(4444)

	// Administration
	config.Log = make(chan string, bufLen)
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

	// Start Logger
	fmt.Println("Starting logger...")
	strongmessage.BlockingLogger(config.Log)

	// Give some time for cleanup...
	fmt.Println("Cleaning up...")
	time.Sleep(time.Millisecond)
}
