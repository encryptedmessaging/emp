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

type ApiConfig struct {
	// Network Channels
	RecvQueue chan quibit.Frame
	SendQueue chan quibit.Frame
	PeerQueue chan quibit.Peer

	// Local Logic
	DbFile       string
	LocalDB      string
	NodeFile     string
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
	RPCUser string
	RPCPass string

	HttpRoot string
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

	RPCConf RPCConf `toml:"rpc"`
}

type RPCConf struct {
	User  string
	Pass  string
	Port  uint16
	Local string `toml:"local_client"`
}

func GetConfDir() string {
	usr, err := user.Current()
	if err != nil {
		return "./"
	}

	return usr.HomeDir + "/.config/emp/"
}

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
	config.PubkeyRegister = make(chan objects.Hash, bufLen)
	config.MessageRegister = make(chan objects.Message, bufLen)
	config.PurgeRegister = make(chan [16]byte, bufLen)

	// Administration
	config.Log = make(chan string, bufLen)
	config.Quit = make(chan os.Signal, 1)

	// Initialize Map
	config.NodeList.Nodes = make(map[string]objects.Node)

	// Bootstrap Nodes
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
