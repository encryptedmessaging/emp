package api

import (
	"fmt"
	"strongmessage/db"
	"quibit"
)

func Start(config *ApiConfig) {
	var err error

	defer quit(config)

	// Start Database Services
	err = db.Initialize(config.Log, config.DbFile)
	defer db.Cleanup()
	if err != nil {
		config.Log <- fmt.Sprintf("Error initializing database: %s", err)
		config.Log <- "Quit"
		return
	}

	err = quibit.Initialize(config.Log, config.RecvQueue, config.SendQueue, config.PeerQueue, config.LocalVersion.Port)
	defer quibit.Cleanup()
	if err != nil {
		config.Log <- fmt.Sprintf("Error initializing network: %s", err)
		return
	}

	for {
		select {
			case <-config.Quit:
				fmt.Println()
				return
		}
	}

	// Should NEVER get here!
	panic("Must've been a cosmic ray!")
}

func quit(config *ApiConfig) {
	config.Log <- "Quit"
}