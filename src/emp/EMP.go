/**
    This file is part of EMP.

    EMP is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
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
