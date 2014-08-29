/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

// Package api provides a TCP server that fully implements the EMProtocol.
package api

import (
	"emp/db"
	"emp/objects"
	"fmt"
	"quibit"
	"runtime"
	"time"
)

// Starts a new TCP Server wth configuration specified in ApiConfig. 
// Server will terminate cleanly only when data is sent to the Quit channel.
// 
// See (struct ApiConfig) for details.
func Start(config *ApiConfig) {
	var err error
	var frame quibit.Frame

	defer quit(config)

	config.Log <- "Starting api..."

	// Start Database Services
	err = db.Initialize(config.Log, config.DbFile)
	defer db.Cleanup()
	if err != nil {
		config.Log <- fmt.Sprintf("Error initializing database: %s", err)
		config.Log <- "Quit"
		return
	}
	config.LocalVersion.Timestamp = time.Now().Round(time.Second)

	locVersion := objects.MakeFrame(objects.VERSION, objects.REQUEST, &config.LocalVersion)
	for str, _ := range config.NodeList.Nodes {
		locVersion.Peer = str
		config.SendQueue <- *locVersion
	}

	// Set Up Clocks
	second := time.Tick(2 * time.Second)
	minute := time.Tick(time.Minute)

	for {
		select {
		case frame = <-config.RecvQueue:
			config.Log <- fmt.Sprintf("Received %s frame...", CmdString(frame.Header.Command))
			switch frame.Header.Command {
			case objects.VERSION:
				version := new(objects.Version)
				err = version.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing version: %s", err)
				} else {
					fVERSION(config, frame, version)
				}
			case objects.PEER:
				nodeList := new(objects.NodeList)
				err = nodeList.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing peer list: %s", err)
				} else {
					fPEER(config, frame, nodeList)
				}
			case objects.OBJ:
				obj := new(objects.Obj)
				err = obj.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing obj list: %s", err)
				} else {
					fOBJ(config, frame, obj)
				}
			case objects.GETOBJ:
				getObj := new(objects.Hash)
				if len(frame.Payload) == 0 {
					break
				}
				err = getObj.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing getobj hash: %s", err)
				} else {
					fGETOBJ(config, frame, getObj)
				}
			case objects.PUBKEY_REQUEST:
				pubReq := new(objects.Hash)
				err = pubReq.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing pubkey request hash: %s", err)
				} else {
					fPUBKEY_REQUEST(config, frame, pubReq)
				}
			case objects.PUBKEY:
				pub := new(objects.EncryptedPubkey)
				err = pub.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing pubkey: %s", err)
				} else {
					fPUBKEY(config, frame, pub)
				}
			case objects.MSG:
				msg := new(objects.Message)
				err = msg.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing message: %s", err)
				} else {
					fMSG(config, frame, msg)
				}
			case objects.PUB:
				msg := new(objects.Message)
				err = msg.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing publication: %s", err)
				} else {
					fPUB(config, frame, msg)
				}
			case objects.PURGE:
				purge := new(objects.Purge)
				err = purge.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing purge: %s", err)
				} else {
					fPURGE(config, frame, purge)
				}
			case objects.CHECKTXID:
				chkTxid := new(objects.Hash)
				if len(frame.Payload) == 0 {
					break
				}
				err = chkTxid.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing checktxid hash: %s", err)
				} else {
					fCHECKTXID(config, frame, chkTxid)
				}
			default:
				config.Log <- fmt.Sprintf("Received invalid frame for command: %d", frame.Header.Command)
			}
		case <-config.Quit:
			fmt.Println()
			// Dump Nodes to File
			DumpNodes(config)
			return
		case <-second:
			// Reconnection Logic
			for key, node := range config.NodeList.Nodes {
				peer := quibit.GetPeer(key)
				if peer == nil || !peer.IsConnected() {
					quibit.KillPeer(key)
					if node.Attempts >= 3 {
						config.Log <- fmt.Sprintf("Max connection attempts reached for %s, disconnecting...", key)
						// Max Attempts Reached, disconnect
						delete(config.NodeList.Nodes, key)
					} else {
						config.Log <- fmt.Sprintf("Disconnected from peer %s, trying to reconnect...", key)
						peer = new(quibit.Peer)
						peer.IP = node.IP
						peer.Port = node.Port
						config.PeerQueue <- *peer
						runtime.Gosched()
						peer = nil
						node.Attempts++
						config.NodeList.Nodes[key] = node
						locVersion.Peer = key
						config.SendQueue <- *locVersion
					}
				}
			}

			if len(config.NodeList.Nodes) < 1 {
				config.Log <- "All connections lost, re-bootstrapping..."

				for i, str := range config.Bootstrap {
					if i >= bufLen {
						break
					}

					p := new(quibit.Peer)
					n := new(objects.Node)
					err := n.FromString(str)
					if err != nil {
						fmt.Println("Error Decoding Peer ", str, ": ", err)
						continue
					}

					p.IP = n.IP
					p.Port = n.Port
					config.PeerQueue <- *p
					runtime.Gosched()
					config.NodeList.Nodes[n.String()] = *n
				}

				for str, _ := range config.NodeList.Nodes {
					locVersion.Peer = str
					config.SendQueue <- *locVersion
				}
			}
		case <-minute:
			// Dump old messages
			err = db.SweepMessages(30 * 24 * time.Hour)
			if err != nil {
				config.Log <- fmt.Sprintf("Error Sweeping Messages: %s", err)
			}
		}
	}

	// Should NEVER get here!
	panic("Must've been a cosmic ray!")
}

func quit(config *ApiConfig) {
	config.Log <- "Quit"
}
