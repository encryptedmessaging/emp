package api

import (
	"os"
	"quibit"
	"strongmessage/objects"
)

type ApiConfig struct {
	// Network Channels
	RecvQueue chan quibit.Frame
	SendQueue chan quibit.Frame
	PeerQueue chan quibit.Peer

	// Local Logic
	DbFile       string
	LocalDB      string
	NodeList     objects.NodeList
	LocalVersion objects.Version

	// Local Register
	PubkeyRegister  chan objects.Hash
	MessageRegister chan objects.Message
	PurgeRegister   chan [16]byte

	// Administration
	Log  chan string
	Quit chan os.Signal

	// Network
	RPCPort uint16
}

func CmdString(cmd uint8) string {
	var ret string

	switch cmd {
	case objects.VERSION:
		ret = "version"
	case objects.PEER:
		ret = "peer list"
	case objects.OBJ:
		ret = "object vector"
	case objects.GETOBJ:
		ret = "object request"
	case objects.PUBKEY_REQUEST:
		ret = "public key request"
	case objects.PUBKEY:
		ret = "public key"
	case objects.MSG:
		ret = "encrypted message"
	case objects.PURGE:
		ret = "purge notification"
	default:
		ret = "unknown"
	}

	return ret
}
