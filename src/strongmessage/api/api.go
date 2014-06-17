package api

import (
	"strongmessage/network"
	"strongmessage/objects"
	"strongmessage/db"
	"crypto/sha512"
	zmq "github.com/alecthomas/gozmq"
	"time"
	"fmt"
)

type ApiConfig struct {
	SendChan  chan network.Frame
	RecvChan  chan network.Frame
	RepRecv   chan network.Frame
	RepSend   chan network.Frame
	PeerChan  chan network.Peer
	Context	  *zmq.Context
	LocalPeer *network.Peer
}

func Start(log chan string, config *ApiConfig, peers network.PeerList) {
	var frame *network.Frame
	var version *objects.Version
	var err error
	version = new(objects.Version)

	// Create local version with which to connect to peers
	localVersion := new(objects.Version)

	localVersion.Version = uint32(objects.LOCAL_VERSION)
	localVersion.UserAgent = objects.LOCAL_USER
	localVersion.Timestamp = time.Now()
	localVersion.IpAddress = config.LocalPeer.IpAddress
	localVersion.Port = config.LocalPeer.Port
	localVersion.AdminPort = config.LocalPeer.AdminPort

	frame = network.NewFrame("version", localVersion.GetBytes(log))
	peers.SendAll(log, frame, config.RecvChan)

	for {
		select {
		case *frame = <-config.RecvChan:
			// Handle Received frames that do not require replies
			if frame.Magic != network.KNOWN_MAGIC {
				continue
			}
			switch frame.Type {
				case "version":
					// Received reply to version request
					log <- fmt.Sprintf("Parsing version reply...")
					err = version.FromBytes(log, frame.Payload)
					if err != nil {
						continue
					}
					go handleVersion(log, config, version, peers)
				case "obj":
					// Received reply to object request
					log <- fmt.Sprintf("Parsing version reply...")
					go handleObj(log, config, frame.Payload, frame.Peer)
				case "peer":
					// Received reply to peer request
					log <- fmt.Sprintf("Parsing peer reply...")
					go handlePeer(log, config, frame.Payload, peers, localVersion, frame.Peer)
				case "pubkeyrq":
					// Received public key request
					log <- fmt.Sprintf("Parsing pubkey request...")
					if len(frame.Payload) != 48 {
						continue
					}
					hashType := db.Contains(string(frame.Payload))
					if hashType == db.NOTFOUND {
						db.Add(string(frame.Payload), db.PUBKEYRQ)
						config.SendChan <- *frame
					} else if hashType == db.PUBKEY {
						pubHash := frame.Payload
						frame.Type = "pubkey"
						frame.Payload, err = db.GetPubkey(log, pubHash)
						if err != nil {
							db.RemoveHash(log, pubHash)
							continue
						}
						config.SendChan <- *frame
					}
				case "pubkey":
					// Received public key request
					log <- fmt.Sprintf("Parsing pubkey...")
					if len(frame.Payload) <= 48 {
						continue
					}
					pubHash := frame.Payload[:48]
					hashType := db.Contains(string(pubHash))
					if hashType == db.PUBKEY {
						continue
					} else {
						db.Delete(string(pubHash))
						err = db.AddPubkey(log, pubHash, frame.Payload[48:])
						config.SendChan <- *frame
					}
				case "msg":
					// Received message
					log <- fmt.Sprintf("Parsing message...")
					msg, err := objects.MessageFromBytes(log, frame.Payload)
					if err != nil {
						continue
					}

					hashType := db.Contains(string(msg.TxidHash))
					if hashType == db.NOTFOUND {
						config.SendChan <- *frame
						err = db.AddMessage(log, &msg)
					}
				case "purge":
					// Received purge request
					log <- fmt.Sprintf("Parsing purge...")
					purgeHash := sha512.Sum384(frame.Payload)
					hashType := db.Contains(string(purgeHash[:]))
					switch hashType {
					case db.MSG:
						db.RemoveHash(log, purgeHash[:])
						fallthrough
					case db.NOTFOUND:
						db.AddPurge(log, purgeHash[:], frame.Payload)
						config.SendChan <- *frame
					}
				default:
					// Received empty getobj or unknown message
					log <- fmt.Sprintf("Received unknown message of type... %s", frame.Type)
			}

		case *frame = <-config.RepRecv:
			// Handle requests that require replies to config.RepSend
		}
	}
}
