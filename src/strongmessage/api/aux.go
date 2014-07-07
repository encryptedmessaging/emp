package api

import (
	"fmt"
	"quibit"
	"strongmessage/db"
	"strongmessage/objects"
	"time"
)

// Handle a Version Request or Reply
func fVERSION(config *ApiConfig, frame quibit.Frame, version *objects.Version) {

	// Verify not objects.BROADCAST
	if frame.Header.Type == objects.BROADCAST {
		// SHUN THE NODE! SHUN IT WITH FIRE!
		config.Log <- "Node sent a version message as a broadcast. Disconnecting..."
		quibit.KillPeer(frame.Peer)
		return
	}

	// Verify Protcol Version, else Disconnect
	if version.Version != objects.LOCAL_VERSION {
		config.Log <- fmt.Sprintf("Peer protocol version does not match local version: %d", version.Version)
		quibit.KillPeer(frame.Peer)
		return
	}

	// Verify Timestamp (5 minute window), else Disconnect
	dur := time.Since(version.Timestamp)
	if dur.Minutes()+5 > 10 {
		config.Log <- fmt.Sprintf("Peer timestamp too far off local time: %s", dur.String())
		quibit.KillPeer(frame.Peer)
		return
	}

	// If backbone node, verify IP
	backbone := false
	for _, b := range []byte(version.IpAddress) {
		if b != 0 {
			backbone = true
		}
	}

	if backbone {
		testIP := quibit.GetPeer(frame.Peer).IP
		if version.IpAddress.String() != testIP.String() {
			config.Log <- fmt.Sprintf("Backbone node broadcast incorrect IP: %s", version.IpAddress.String())
			quibit.KillPeer(frame.Peer)
			return
		}

		// Add to Master Node List
		var node objects.Node

		node.IP = version.IpAddress
		node.Port = version.Port
		node.LastSeen = time.Now().Round(time.Second)
		config.NodeList.Nodes[node.String()] = node
	}

	var sending *quibit.Frame
	if frame.Header.Type == objects.REQUEST {
		// If a objects.REQUEST, send local version as a objects.REPLY
		sending = objects.MakeFrame(objects.VERSION, objects.REPLY, &config.LocalVersion)
	} else {
		// If a objects.REPLY, send a peer list as a objects.REQUEST
		sending = objects.MakeFrame(objects.PEER, objects.REQUEST, &config.NodeList)
	}
	sending.Peer = frame.Peer
	config.SendQueue <- *sending
} // End fVERSION

// Handle Peer List Requests or Replies
func fPEER(config *ApiConfig, frame quibit.Frame, nodeList *objects.NodeList) {

	// Verify not objects.BROADCAST
	if frame.Header.Type == objects.BROADCAST {
		// SHUN THE NODE! SHUN IT WITH FIRE!
		config.Log <- "Node sent a peer frame as a broadcast. Disconnecting..."
		quibit.KillPeer(frame.Peer)
		return
	}

	var sending *quibit.Frame
	if frame.Header.Type == objects.REQUEST {
		// If a objects.REQUEST, send back peer objects.REPLY
		sending = objects.MakeFrame(objects.PEER, objects.REPLY, &config.NodeList)
	} else {
		// If a objects.REPLY, send an object list as a objects.REQUEST
		sending = objects.MakeFrame(objects.OBJ, objects.REQUEST, db.ObjList())
	}

	sending.Peer = frame.Peer

	config.SendQueue <- *sending

	// Merge incoming list with current list
	for key, node := range nodeList.Nodes {
		_, ok := config.NodeList.Nodes[key]
		if !ok {
			config.NodeList.Nodes[key] = node
			p := new(quibit.Peer)
			p.IP = node.IP
			p.Port = node.Port
			config.PeerQueue <- *p
			time.Sleep(time.Millisecond)
			newVer := objects.MakeFrame(objects.VERSION, objects.REQUEST, &config.LocalVersion)
			config.SendQueue <- *newVer
		} // End if
	} // End for
} // End fPEER

// Handle Object Vector Requests or Replies
func fOBJ(config *ApiConfig, frame quibit.Frame, obj *objects.Obj) {
	var sending *quibit.Frame

	// Verify not objects.BROADCAST
	if frame.Header.Type == objects.BROADCAST {
		// SHUN THE NODE! SHUN IT WITH FIRE!
		config.Log <- "Node sent an obj frame as a broadcast. Disconnecting..."
		quibit.KillPeer(frame.Peer)
		return
	}

	if frame.Header.Type == objects.REQUEST {
		// If a objects.REQUEST, send local object list as objects.REPLY
		sending = objects.MakeFrame(objects.OBJ, objects.REPLY, db.ObjList())
		sending.Peer = frame.Peer
		config.SendQueue <- *sending
	}

	// For each object in object list:
	// If object not stored locally, send GETOBJ objects.REQUEST
	for _, hash := range obj.HashList {
		if db.Contains(hash) == db.NOTFOUND {
			sending = objects.MakeFrame(objects.GETOBJ, objects.REQUEST, &hash)
			sending.Peer = frame.Peer
			config.SendQueue <- *sending
		}
	}
}

// Handle Object Detail Requests
func fGETOBJ(config *ApiConfig, frame quibit.Frame, hash *objects.Hash) {
	// Verify not objects.BROADCAST
	if frame.Header.Type == objects.BROADCAST {
		// SHUN THE NODE! SHUN IT WITH FIRE!
		config.Log <- "Node sent a getobj message as a broadcast. Disconnecting..."
		quibit.KillPeer(frame.Peer)
		return
	}

	// If object stored locally, send object as a objects.REPLY
	var sending *quibit.Frame
	if frame.Header.Type == objects.REQUEST {
		switch db.Contains(*hash) {
		case db.PUBKEY:
			sending = objects.MakeFrame(objects.PUBKEY, objects.REPLY, db.GetPubkey(config.Log, *hash))
		case db.PURGE:
			sending = objects.MakeFrame(objects.PURGE, objects.REPLY, db.GetPurge(config.Log, *hash))
		case db.MSG:
			sending = objects.MakeFrame(objects.MSG, objects.REPLY, db.GetMessage(config.Log, *hash))
		case db.PUBKEYRQ:
			sending = objects.MakeFrame(objects.PUBKEY_REQUEST, objects.REPLY, hash)
		default:
			sending = objects.MakeFrame(objects.GETOBJ, objects.REPLY, new(objects.NilPayload))
		} // End switch
		sending.Peer = frame.Peer
		config.SendQueue <- *sending
	} // End if
} // End fGETOBJ

// Handle Public Key Request Broadcasts
func fPUBKEY_REQUEST(config *ApiConfig, frame quibit.Frame, pubHash *objects.Hash) {
	// Check Hash in Object List

	switch db.Contains(*pubHash) {
	// If request is Not in List, store the request
	case db.NOTFOUND:
		// If a objects.BROADCAST, send out another objects.BROADCAST
		db.Add(*pubHash, db.PUBKEYRQ)
		if frame.Header.Type == objects.BROADCAST {
			config.SendQueue <- *objects.MakeFrame(objects.PUBKEY_REQUEST, objects.BROADCAST, pubHash)
		}

	// If request is a Public Key in List:
	case db.PUBKEY:
		// Send out the PUBKEY as a objects.BROADCAST
		config.SendQueue <- *objects.MakeFrame(objects.PUBKEY, objects.BROADCAST, db.GetPubkey(config.Log, *pubHash))
	}
}

// Handle Public Key Broadcasts
func fPUBKEY(config *ApiConfig, frame quibit.Frame, pubkey *objects.EncryptedPubkey) {
	// Check Hash in Object List
	switch db.Contains(pubkey.AddrHash) {
	// If request is a Pubkey Request, remove the pubkey request
	case db.PUBKEYRQ:
		db.Delete(pubkey.AddrHash)
		fallthrough
	case db.NOTFOUND:
		// Add Pubkey to database
		err := db.AddPubkey(config.Log, *pubkey)
		if err != nil {
			config.Log <- fmt.Sprintf("Error adding pubkey to database: %s", err)
			break
		}
		// If a objects.BROADCAST, send a objects.BROADCAST
		if frame.Header.Type == objects.BROADCAST {
			config.SendQueue <- *objects.MakeFrame(objects.PUBKEY, objects.BROADCAST, pubkey)
		}

		config.PubkeyRegister <- pubkey.AddrHash
	}
} // End fPUBKEY

// Handle Encrypted Message Broadcasts
func fMSG(config *ApiConfig, frame quibit.Frame, msg *objects.Message) {
	// Check Hash in Object List
	switch db.Contains(msg.TxidHash) {
	// If Not in List, Store and objects.BROADCAST
	case db.NOTFOUND:
		err := db.AddMessage(config.Log, msg)
		if err != nil {
			config.Log <- fmt.Sprintf("Error adding message to database: %s", err)
			break
		}
		if frame.Header.Type == objects.BROADCAST {
			config.SendQueue <- *objects.MakeFrame(objects.MSG, objects.BROADCAST, msg)
		}

		config.MessageRegister <- *msg

	// If found as PURGE, reply with PURGE
	case db.PURGE:
		sending := objects.MakeFrame(objects.PURGE, objects.REPLY, db.GetPurge(config.Log, msg.TxidHash))
		sending.Peer = frame.Peer
		config.SendQueue <- *sending
	}
} // End fMSG

// Handle Purge Broadcasts
func fPURGE(config *ApiConfig, frame quibit.Frame, purge *objects.Purge) {
	var err error
	txidHash := objects.MakeHash(purge.Txid[:])

	// Check Hash in Object List
	switch db.Contains(txidHash) {
	// Delete Stored Messages
	case db.MSG:
		err = db.RemoveHash(config.Log, txidHash)
		if err != nil {
			config.Log <- fmt.Sprintf("Error removing message from database: %s", err)
			break
		}
		fallthrough
	// Add to database
	case db.NOTFOUND:
		err = db.AddPurge(config.Log, *purge)
		if err != nil {
			config.Log <- fmt.Sprintf("Error adding purge to database: ", err)
			break
		}

		// Re-objects.BROADCAST if necessary
		if frame.Header.Type == objects.BROADCAST {
			config.SendQueue <- *objects.MakeFrame(objects.PURGE, objects.BROADCAST, purge)
		}
		config.PurgeRegister <- purge.Txid
	} // End Switch
} // End fPURGE
