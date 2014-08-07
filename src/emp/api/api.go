package api

import (
	"emp/db"
	"emp/objects"
	"fmt"
	"quibit"
	"time"
)

func Start(config *ApiConfig) {
	var err error
	var frame quibit.Frame

	defer quit(config)

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
			case objects.CHANNEL:
				msg := new(objects.Message)
				err = msg.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing channel message: %s", err)
				} else {
					fCHANNEL(config, frame, msg)
				}
			default:
				config.Log <- fmt.Sprintf("Received invalid frame for command: %d", frame.Header.Command)
			}
		case <-config.Quit:
			fmt.Println()
			// Dump Nodes to File
			DumpNodes(config)
			return
		}
	}

	// Should NEVER get here!
	panic("Must've been a cosmic ray!")
}

func quit(config *ApiConfig) {
	config.Log <- "Quit"
}
