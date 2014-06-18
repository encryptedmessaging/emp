package api

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"net"
	"strongmessage/network"
	"strongmessage/objects"
	"strongmessage/db"
	"testing"
	"time"
	"crypto/sha512"
	"os/exec"
)

func setup() (chan string, *ApiConfig, *network.PeerList) {
	config := new(ApiConfig)

	config.Context, _ = zmq.NewContext()

	config.LocalPeer = new(network.Peer)
	config.LocalPeer.IpAddress = net.ParseIP("127.0.0.1")
	config.LocalPeer.Port = 4444
	config.LocalPeer.AdminPort = 4445

	config.SendChan = make(chan network.Frame, 10)
	config.RecvChan = make(chan network.Frame, 10)
	config.RepRecv = make(chan network.Frame, 10)
	config.RepSend = make(chan network.Frame, 10)
	config.PeerChan = make(chan network.Peer, 10)
	config.DBFile = "testdb.db"

	log := make(chan string, 100)

	peers := new(network.PeerList)

	go Start(log, config, peers)

	return log, config, peers
}

func cleanup(config *ApiConfig, peers *network.PeerList) {
	peers.DisconnectAll()
	config.Context.Close()

	db.Cleanup()
	exec.Command("rm", "testdb.db").Run()
}

func TestVersionPeer(t *testing.T) {
	// Setup Environment
	log, config, peers := setup()

	frame := new(network.Frame)
	var err error

	// Test 1: Version Messages
	version := new(objects.Version)
	version.Version = uint32(1)
	version.Timestamp = time.Now()
	version.IpAddress = net.ParseIP("127.0.0.1")
	version.Port = 4444
	version.AdminPort = 4445
	version.UserAgent = objects.LOCAL_USER

	config.RepRecv <- *network.NewFrame("version", version.GetBytes(log))

	*frame = <-config.RepSend

	if frame.Type != "version" {
		fmt.Println("Error: version type incorrect")
		t.FailNow()
	}

	err = version.FromBytes(log, frame.Payload)
	if err != nil {
		fmt.Println("Error: Cannot parse version...", frame.Payload)
		t.FailNow()
	}

	if version.Version != objects.LOCAL_VERSION || version.IpAddress.String() != "127.0.0.1" || version.Port != 4444 || version.AdminPort != 4445 || version.UserAgent != objects.LOCAL_USER {
		fmt.Println("Error: bytes of responded version are incorrect...", *version)
		t.FailNow()
	}

	// Test 2: Peer Requests
	testPeer := new(network.Peer)
	testPeer2 := new(network.Peer)

	testPeer.IpAddress = net.ParseIP("0.0.0.1")
	testPeer.Port = 4444
	testPeer.AdminPort = 4445
	testPeer.LastSeen = time.Now()

	config.RepRecv <- *network.NewFrame("peer", testPeer.GetBytes())

	*frame = <-config.RepSend

	if frame.Type != "peer" {
		fmt.Println("Error: peer type incorrect")
		t.FailNow()
	}

	// Response should be exactly 1 peer
	err = testPeer2.FromBytes(frame.Payload)

	if err != nil {
		fmt.Println("Error: could not parse peer from peer response...", frame.Payload)
		t.FailNow()
	}

	if testPeer2.IpAddress.String() != "127.0.0.1" || testPeer2.Port != 4444 || testPeer2.AdminPort != 4445 {
		fmt.Println("Error: Peer response is incorrect... ", testPeer2)
		t.FailNow()
	}

	*testPeer2 = <-config.PeerChan

	// Should match the version request from earlier
	if testPeer2.IpAddress.String() != "127.0.0.1" || testPeer2.Port != 4444 || testPeer2.AdminPort != 4445 {
		fmt.Println("Error: peer sent to peerChan doesn't match!", testPeer2.GetBytes(), testPeer.GetBytes())
		t.FailNow()
	}

	cleanup(config, peers)
}

func TestPubkey(t *testing.T) {
	_, config, peers := setup()

	addrHash := sha512.Sum384([]byte{'a', 'b', 'c', 'd'})
	var pubkey [65]byte
	for i:=0; i<65; i++ {
		pubkey[i] = 1
	}

	var frame network.Frame

	// Test 3: Public Key Request
	config.RecvChan <- *network.NewFrame("pubkeyrq", addrHash[:])

	frame = <-config.SendChan

	if frame.Type != "pubkeyrq" {
		fmt.Println("Error: pubkeyrq type incorrect")
		t.FailNow()
	}

	if string(frame.Payload) != string(addrHash[:]) {
		fmt.Println("Error: Incorrect test 3 payload: ", frame.Payload)
		t.FailNow()
	}

	// Test 4: Tested Later, should not send out another pubkeyrq
	config.RecvChan <- *network.NewFrame("pubkeyrq", addrHash[:])

	// Test 5: Public Key
	config.RecvChan <- *network.NewFrame("pubkey", append(addrHash[:], pubkey[:]...))

	frame = <-config.SendChan

	if frame.Type != "pubkey" {
		fmt.Println("Error: pubkeyrq was resent!")
		t.FailNow()
	}

	if string(frame.Payload) != string(append(addrHash[:], pubkey[:]...)) {
		fmt.Println("Error: Incorrect test 5 payload: ", frame.Payload, append(addrHash[:], pubkey[:]...))
		t.FailNow()
	}

	// Test 6: Public Key Request returns Pubkey
	config.RecvChan <- *network.NewFrame("pubkeyrq", addrHash[:])

	frame = <-config.SendChan

	if frame.Type != "pubkey" {
		fmt.Println("Error: public key not sent back! Type: ", frame.Type)
		t.FailNow()
	}

	if string(frame.Payload) != string(append(addrHash[:], pubkey[:]...)) {
		fmt.Println("Error: Incorrect test 6 payload: ", frame.Payload, append(addrHash[:], pubkey[:]...))
		t.FailNow()
	}

	cleanup(config, peers)
}