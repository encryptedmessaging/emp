package api

import (
	"fmt"
	"quibit"
	"strongmessage/db"
	"strongmessage/objects"
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

	err = quibit.Initialize(config.Log, config.RecvQueue, config.SendQueue, config.PeerQueue, config.LocalVersion.Port)
	defer quibit.Cleanup()
	if err != nil {
		config.Log <- fmt.Sprintf("Error initializing network: %s", err)
		return
	}

	for {
		select {
		case frame = <-config.RecvQueue:
			config.Log <- fmt.Sprintf("Received %s frame...", CmdString(frame.Header.Command))
			switch frame.Header.Command {
			case VERSION:
				version := new(objects.Version)
				err = version.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing version: %s", err)
				} else {
					go fVERSION(config, frame, version)
				}
			case PEER:
				nodeList := new(objects.NodeList)
				err = nodeList.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing peer list: %s", err)
				} else {
					go fPEER(config, frame, nodeList)
				}
			case OBJ:
				obj := new(objects.Obj)
				err = obj.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing obj list: %s", err)
				} else {
					go fOBJ(config, frame, obj)
				}
			case GETOBJ:
				getObj := new(objects.Hash)
				err = getObj.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing getobj hash: %s", err)
				} else {
					go fGETOBJ(config, frame, getObj)
				}
			case PUBKEY_REQUEST:
				pubReq := new(objects.Hash)
				err = pubReq.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing pubkey request hash: %s", err)
				} else {
					go fPUBKEY_REQUEST(config, frame, pubReq)
				}
			case PUBKEY:
				pub := new(objects.EncryptedPubkey)
				err = pub.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing pubkey: %s", err)
				} else {
					go fPUBKEY(config, frame, pub)
				}
			case MSG:
				msg := new(objects.Message)
				err = msg.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing message: %s", err)
				} else {
					go fMSG(config, frame, msg)
				}
			case PURGE:
				purge := new(objects.Purge)
				err = purge.FromBytes(frame.Payload)
				if err != nil {
					config.Log <- fmt.Sprintf("Error parsing purge: %s", err)
				} else {
					go fPURGE(config, frame, purge)
				}
			default:
				config.Log <- fmt.Sprintf("Received invalid frame for command: %d", frame.Header.Command)
			}
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
