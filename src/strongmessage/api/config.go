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
	NodeList     objects.NodeList
	LocalVersion objects.Version

	// Administration
	Log  chan string
	Quit chan os.Signal
}

// Message Commands
const (
	VERSION = iota
	PEER    = iota
	OBJ     = iota
	GETOBJ  = iota

	PUBKEY_REQUEST = iota
	PUBKEY         = iota
	MSG            = iota
	PURGE          = iota

	SHUN = iota
)

// Message Types
const (
	BROADCAST = iota
	REQUEST   = iota
	REPLY     = iota
)

func CmdString(cmd uint8) string {
	var ret string

	switch cmd {
	case VERSION:
		ret = "version"
	case PEER:
		ret = "peer list"
	case OBJ:
		ret = "object vector"
	case GETOBJ:
		ret = "object request"
	case PUBKEY_REQUEST:
		ret = "public key request"
	case PUBKEY:
		ret = "public key"
	case MSG:
		ret = "encrypted message"
	case PURGE:
		ret = "purge notification"
	default:
		ret = "unknown"
	}

	return ret
}
