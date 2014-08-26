/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package api

import (
	"emp/objects"
	"fmt"
	"os"
	"os/exec"
	"quibit"
	"testing"
	"time"
)

func initialize() *ApiConfig {
	config := new(ApiConfig)

	// Network Channels
	config.RecvQueue = make(chan quibit.Frame)
	config.SendQueue = make(chan quibit.Frame)
	config.PeerQueue = make(chan quibit.Peer)

	// Local Logic
	config.DbFile = "testdb.db"

	config.LocalVersion.Version = 1
	config.LocalVersion.Timestamp = time.Now().Round(time.Second)
	config.LocalVersion.Port = 4444
	config.LocalVersion.UserAgent = "strongmsg v0.1"

	// Administration
	config.Log = make(chan string, 100)
	config.Quit = make(chan os.Signal, 1)

	go Start(config)

	return config
}

func cleanup(config *ApiConfig) {
	var s os.Signal
	config.Quit <- s

	str := <-config.Log
	for str != "Quit" {
		fmt.Println(str)
		str = <-config.Log
	}

	exec.Command("rm", "testdb.db").Run()

}

func TestHandshake(t *testing.T) {
	config := initialize()

	var frame quibit.Frame
	var err error

	// Test Version
	frame = *objects.MakeFrame(objects.VERSION, objects.REQUEST, &config.LocalVersion)
	frame.Peer = "127.0.0.1:4444"

	config.RecvQueue <- frame

	frame = <-config.SendQueue

	if frame.Header.Command != objects.VERSION || frame.Header.Type != objects.REPLY {
		fmt.Println("Frame is not a proper reply to a version request: ", frame.Header)
		t.FailNow()
	}

	version := new(objects.Version)
	err = version.FromBytes(frame.Payload)
	if err != nil {
		fmt.Println("Error parsing version reply: ", err)
		t.FailNow()
	}

	// Test Peer
	frame = *objects.MakeFrame(objects.PEER, objects.REQUEST, &config.NodeList)
	frame.Peer = "127.0.0.1:4444"

	config.RecvQueue <- frame

	frame = <-config.SendQueue

	if frame.Header.Command != objects.PEER || frame.Header.Type != objects.REPLY || frame.Header.Length != 0 {
		fmt.Println("Frame is not a proper reply to a peer request: ", frame.Header)
		t.FailNow()
	}

	// Test Obj
	frame = *objects.MakeFrame(objects.OBJ, objects.REQUEST, &config.NodeList)
	frame.Peer = "127.0.0.1:4444"

	config.RecvQueue <- frame

	frame = <-config.SendQueue

	if frame.Header.Command != objects.OBJ || frame.Header.Type != objects.REPLY || frame.Header.Length != 0 {
		fmt.Println("Frame is not a proper reply to a peer request: ", frame.Header)
		t.FailNow()
	}

	cleanup(config)
}
