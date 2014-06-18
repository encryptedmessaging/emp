package api

import (
	"crypto/sha512"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"strongmessage/db"
	"strongmessage/network"
	"strongmessage/objects"
	"time"
)

type ApiConfig struct {
	SendChan  chan network.Frame
	RecvChan  chan network.Frame
	RepRecv   chan network.Frame
	RepSend   chan network.Frame
	PeerChan  chan network.Peer
	Context   *zmq.Context
	LocalPeer *network.Peer
}

func Start(log chan string, config *ApiConfig, peers *network.PeerList) {
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
			if frame.Magic != network.KNOWN_MAGIC {
				continue
			}
			switch frame.Type {
			case "version":
				// Received version request
				log <- fmt.Sprintf("Parsing version request...")
				err = version.FromBytes(log, frame.Payload)
				if err != nil {
					continue
				}
				newPeer := new(network.Peer)

				// Setup and Send Response
				frame.Payload = localVersion.GetBytes(log)

				newPeer.IpAddress = version.IpAddress
				newPeer.Port = version.Port
				newPeer.AdminPort = version.AdminPort
				newPeer.LastSeen = time.Now()
				for _, peer := range peers.Peers {
					if peer.TcpString() == newPeer.TcpString() {
						peer.LastSeen = newPeer.LastSeen
						log <- fmt.Sprintf("Peer %s updated", peer.TcpString())
						continue
					}
				}

				newPeer.Connect(log, config.Context)
				peers.Peers = append(peers.Peers, *newPeer)
				config.PeerChan <- *newPeer
				config.RepSend <- *frame

			case "peer":
				// Received peer request
				log <- fmt.Sprintf("Parsing peer request...")
				newPeers := frame.Payload
				// Condense PeerList to Peer Hash Table
				peerHash := make(map[string]*network.Peer)

				for _, peer := range peers.Peers {
					peerHash[peer.TcpString()] = &peer
				}

				for i := 0; i <= len(newPeers)-28; i += 28 {
					p := new(network.Peer)
					err := p.FromBytes(newPeers[i : i+28])
					if err != nil {
						continue
					}

					_, ok := peerHash[p.TcpString()]
					if !ok {

						// Send Version Request
						err = p.Connect(log, config.Context)
						peers.Peers = append(peers.Peers, *p)
						if err != nil {
							p.SendRequest(log, network.NewFrame("version", localVersion.GetBytes(log)), config.RecvChan)
						}

					}
				}

				payload := make([]byte, 0, 0)
				for _, peer := range peerHash {
					payload = append(payload, peer.GetBytes()...)
				}
				config.RepSend <- *network.NewFrame("peer", payload)
			case "obj":
				// Received object request
				hashes := db.HashCopy()

				objs := frame.Payload

				for i := 0; i <= len(objs)-48; i += 48 {
					if db.Contains(string(objs[i:i+48])) != db.NOTFOUND {
						delete(hashes, string(objs[i:i+48]))
					} else {
						for j := 1; j < 5 && j <= len(peers.Peers); j++ {
							peers.Peers[len(peers.Peers)-j].SendRequest(log, network.NewFrame("getobj", objs[i:i+48]), config.RecvChan)
						}
					}
				}

				payload := make([]byte, 0, 0)

				for hash, _ := range hashes {
					payload = append(payload, []byte(hash)...)
				}

				config.RepSend <- *network.NewFrame("obj", payload)
			case "getobj":
				// Received object request
				hash := frame.Payload

				switch db.Contains(string(hash)) {
				case db.MSG:
					msg, err := db.GetMessage(log, hash)
					if err != nil {
						config.RepSend <- *network.NewFrame("getobj", nil)
					} else {
						config.RepSend <- *network.NewFrame("msg", msg.GetBytes(log))
					}
				case db.PUBKEY:
					pub, err := db.GetPubkey(log, hash)
					if err != nil {
						config.RepSend <- *network.NewFrame("getobj", nil)
					} else {
						config.RepSend <- *network.NewFrame("pubkey", append(hash, pub...))
					}
				case db.PUBKEYRQ:
					config.RepSend <- *network.NewFrame("pubkeyrq", hash)
				case db.PURGE:
					pur, err := db.GetPurge(log, hash)
					if err != nil {
						config.RepSend <- *network.NewFrame("getobj", nil)
					} else {
						config.RepSend <- *network.NewFrame("purge", pur)
					}
				default:
					config.RepSend <- *network.NewFrame("getobj", nil)
				}

			default:
				log <- fmt.Sprintf("Received unknown message of type... %s", frame.Type)
			}
		}
	}
}
