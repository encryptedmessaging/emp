package api

import (
	"quibit"
	"strongmessage/objects"
	"os"
)

type ApiConfig struct {
	// Network Channels
	RecvQueue chan quibit.Frame
	SendQueue chan quibit.Frame
	PeerQueue chan quibit.Peer

	// Local Logic
	DbFile string
	NodeList objects.NodeList
	LocalVersion objects.Version

	// Administration
	Log chan string
	Quit chan os.Signal
}

// Message Commands
const (
	VERSION = iota
	PEER = iota
	OBJ = iota
	GETOBJ = iota

	PUBKEY_REQUEST = iota
	PUBKEY = iota
	MSG = iota
	PURGE = iota

	SHUN = iota
)

// Message Types
const (
	BROADCAST = iota
	REQUEST = iota
	REPLY = iota
)
