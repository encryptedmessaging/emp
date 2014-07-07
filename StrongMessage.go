package main

import (
	"fmt"
	"os"
	"os/signal"
	"quibit"
	"strongmessage"
	"strongmessage/api"
	"strongmessage/local/localapi"
)

func main() {

	confFile := api.GetConfDir() + "msg.conf"

	config := api.GetConfig(confFile)

	if config == nil {
		fmt.Println("Error Loading Config, exiting...")
		return
	}

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
