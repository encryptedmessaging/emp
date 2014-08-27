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
	"github.com/BurntSushi/toml"
	"net"
	"os"
	"bufio"
	"os/user"
	"quibit"
	"time"
	"io/ioutil"
)

var confDir string;

type ApiConfig struct {
	// Network Channels
	RecvQueue chan quibit.Frame // Send frames here to be handled by the Running API
	SendQueue chan quibit.Frame // Frames to be broadcast to the network are sent here
	PeerQueue chan quibit.Peer // New peers to connect to are sent here

	// Local Logic
	DbFile       string // Inventory File relative to Config Directory
	LocalDB      string // EMPLocal Database relative to Config Directory
	NodeFile     string // File to store list of <IP>:<Host> Strings
	NodeList     objects.NodeList // Active list of connected backbone nodes.
	LocalVersion objects.Version // Local version broadcast to nodes upon connection
	Bootstrap    []string // List of bootstrap nodes to use when all other nodes are disconnected.

	// Local Register
	PubkeyRegister  chan objects.Hash // Identifiers for incoming encrypted public keys are sent here.
	MessageRegister chan objects.Message // Incoming basic messages are copied here.
	PubRegister     chan objects.Message // Incoming published messages are copied here.
	PurgeRegister   chan [16]byte // Incomping purge tokens are copied here.

	// Administration
	Log  chan string // Error messages are sent here. WILL BLOCK if not 
	Quit chan os.Signal // Send data here to cleanly quit the API Server

	// Network
	RPCPort uint16 // Port to run RPC API and EMPLocal Client
	RPCUser string // Username for RPC server
	RPCPass string // Password for RPC Server

	HttpRoot string // HTML Root of EMPLocal Client
}

// Returns Human-Readable string for a specific EMP command.
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
	case objects.PUB:
		ret = "encrypted publication"
	case objects.PURGE:
		ret = "purge notification"
	case objects.CHECKTXID:
		ret = "purge check"
	default:
		ret = "unknown"
	}

	return ret
}

const (
	bufLen = 10
)

type tomlConfig struct {
	Inventory string `toml:"inventory"`
	Local     string `toml:"local"`
	Nodes     string `toml:"nodes"`

	IP   string
	Port uint16

	Peers []string `toml:"bootstrap"`

	RPCConf rpcConf `toml:"rpc"`
}

type rpcConf struct {
	User  string
	Pass  string
	Port  uint16
	Local string `toml:"local_client"`
}

// Set Config Directory where databases and configuration are stored.
func SetConfDir(conf string) {
	confDir = conf
}

// Get Config Directory: Defaults to $(HOME)/.config/emp/
func GetConfDir() string {
	if len(confDir) != 0 {
		return confDir
	}

	usr, err := user.Current()
	if err != nil {
		return "./"
	}

	return usr.HomeDir + "/.config/emp/"
}

// Generate new config from configuration file. File provided as an Absolute Path.
func GetConfig(confFile string) *ApiConfig {

	var tomlConf tomlConfig

	if _, err := toml.DecodeFile(confFile, &tomlConf); err != nil {
		fmt.Println("Config Error: ", err)
		return nil
	}

	config := new(ApiConfig)

	// Network Channels
	config.RecvQueue = make(chan quibit.Frame, bufLen)
	config.SendQueue = make(chan quibit.Frame, bufLen)
	config.PeerQueue = make(chan quibit.Peer, bufLen)

	// Local Logic
	config.DbFile = GetConfDir() + tomlConf.Inventory
	config.LocalDB = GetConfDir() + tomlConf.Local
	if len(config.DbFile) == 0 || len(config.LocalDB) == 0 {
		fmt.Println("Database file not found in config!")
		return nil
	}
	config.NodeFile = GetConfDir() + tomlConf.Nodes

	config.LocalVersion.Port = tomlConf.Port
	if tomlConf.IP != "0.0.0.0" {
		config.LocalVersion.IpAddress = net.ParseIP(tomlConf.IP)
	}
	config.LocalVersion.Timestamp = time.Now().Round(time.Second)
	config.LocalVersion.Version = objects.LOCAL_VERSION
	config.LocalVersion.UserAgent = objects.LOCAL_USER

	// RPC
	config.RPCPort = tomlConf.RPCConf.Port
	config.RPCUser = tomlConf.RPCConf.User
	config.RPCPass = tomlConf.RPCConf.Pass
	config.HttpRoot = GetConfDir() + tomlConf.RPCConf.Local

	// Local Registers
	config.PubkeyRegister  = make(chan objects.Hash, bufLen)
	config.MessageRegister = make(chan objects.Message, bufLen)
	config.PubRegister     = make(chan objects.Message, bufLen)
	config.PurgeRegister   = make(chan [16]byte, bufLen)

	// Administration
	config.Log = make(chan string, bufLen)
	config.Quit = make(chan os.Signal, 1)

	// Initialize Map
	config.NodeList.Nodes = make(map[string]objects.Node)

	// Bootstrap Nodes
	config.Bootstrap = make([]string, len(tomlConf.Peers), cap(tomlConf.Peers))
	copy(config.Bootstrap, tomlConf.Peers)
	for i, str := range tomlConf.Peers {
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
		config.NodeList.Nodes[n.String()] = *n
	}

	// Pull Nodes from node file
	if len(config.NodeFile) > 0 {
		ReadNodes(config)
	}

	return config
}

// Load and connect to all nodes from the NodeFile found in the ApiConfig.
func ReadNodes(config *ApiConfig) {
	file, err := os.Open(config.NodeFile);
	defer file.Close()
	if err != nil {
		fmt.Println("Could not open node file: ", err)
	}

	var count int

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
    	str := scanner.Text()
    	if len(str) < 0 || str == "<nil>" {
    		continue
    	}

		p := new(quibit.Peer)
		n := new(objects.Node)
		err = n.FromString(str)
		if err != nil {
			fmt.Println("Error Decoding Peer ", str, ": ", err)
			continue
		}

		p.IP = n.IP
		p.Port = n.Port
		config.PeerQueue <- *p
		config.NodeList.Nodes[n.String()] = *n
		count++
	}
	fmt.Println(count, "nodes pulled from node file.")
}

// Dump all nodes in config.NodeList to config.NodeFile.
func DumpNodes(config *ApiConfig) {
	if config == nil {
		return
	}
	if len(config.NodeFile) < 1 {
		return
	}
	writeBytes := make([]byte, 0, 0)

	for key, _ := range config.NodeList.Nodes {
		if quibit.GetPeer(key).IsConnected() {
			writeBytes = append(writeBytes, key...)
			writeBytes = append(writeBytes, byte('\n'))
		}
	}

	err := ioutil.WriteFile(config.NodeFile, writeBytes, 0644)
	if err != nil {
		fmt.Println("Error writing peers to file: ", err)
	}
}
