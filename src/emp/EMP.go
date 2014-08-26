/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package main

import (
	"emp/api"
	"emp/local/localapi"
	"fmt"
	"os"
	"os/signal"
	"quibit"
)

func BlockingLogger(channel chan string) {
        var log string
        for {
                log = <-channel
                fmt.Println(log)
                if log == "Quit" {
                        break
                }
        }
}

func main() {

	if len(os.Args) > 2 {
		fmt.Println("Usage: emp [config_directory]")
		return
	}

	if len(os.Args) == 2 {
		api.SetConfDir(os.Args[1])
	}

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
		fmt.Printf("Error initializing network: %s", err)
		return
	}

	// Start Signal Handler
	signal.Notify(config.Quit, os.Interrupt, os.Kill)

	// Start API
	go api.Start(config)

	go localapi.Initialize(config)

	// Start Logger
	fmt.Println("Starting logger...")
	BlockingLogger(config.Log)

	// Give some time for cleanup...
	fmt.Println("Cleaning up...")
	localapi.Cleanup()
}
