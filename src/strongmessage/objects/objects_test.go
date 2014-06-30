package objects

import (
	"crypto/elliptic"
	"fmt"
	"net"
	"strongmessage/encryption"
	"testing"
	"time"
)

func TestVersion(t *testing.T) {
	v := new(Version)
	v.Version = uint16(1)
	v.Timestamp = time.Unix(0, 0)
	v.IpAddress = net.ParseIP("1.2.3.4")
	v.Port = uint16(4444)
	v.UserAgent = "Hello World!"

	verBytes := v.GetBytes()
	if len(verBytes) != verLen+12 {
		fmt.Println("Incorrect Byte Length: ", verBytes)
		t.Fail()
	}

	v2 := new(Version)
	err := v2.FromBytes(verBytes)
	if err != nil {
		fmt.Println("Error Decoding: ", err)
		t.FailNow()
	}

	if v2.Version != 1 || v2.Timestamp != time.Unix(0, 0) || v2.IpAddress.String() != "1.2.3.4" || v2.Port != 4444 || v2.UserAgent != "Hello World!" {
		fmt.Println("Incorrect decoded version: ", v2)
		t.Fail()
	}
}

func TestNodes(t *testing.T) {
	n := new(NodeList)
	n2 := new(NodeList)
	node1 := new(Node)
	node2 := new(Node)

	node1.IP = net.ParseIP("1.2.3.4")
	node1.Port = uint16(4444)
	node1.LastSeen = time.Now().Round(time.Second)

	node2.IP = net.ParseIP("5.6.7.8")
	node2.Port = uint16(5555)
	node2.LastSeen = time.Now().Round(time.Second)

	n.Nodes = append(n.Nodes, *node1)
	n.Nodes = append(n.Nodes, *node2)

	nBytes := n.GetBytes()
	if len(nBytes) != 2*nodeLen {
		fmt.Println("Byte Lengh Incorrect: ", nBytes)
		t.FailNow()
	}

	err := n2.FromBytes(nBytes)
	if err != nil {
		fmt.Println("Error Decoding: ", err)
	}

	if n2.Nodes[0].IP.String() != n.Nodes[0].IP.String() || n2.Nodes[1].IP.String() != n.Nodes[1].IP.String() || n2.Nodes[0].Port != n.Nodes[0].Port || n2.Nodes[1].Port != n.Nodes[1].Port {
		fmt.Println("Nodes don't match!", n2.Nodes)
		t.Fail()
	}
}

func TestObj(t *testing.T) {
	o := new(Obj)
	o.HashList = make([]Hash, 2, 2)

	o.HashList[0] = MakeHash([]byte{'a', 'b', 'c', 'd'})
	o.HashList[1] = MakeHash([]byte{'e', 'f', 'g', 'h'})

	o2 := new(Obj)

	oBytes := o.GetBytes()
	if len(oBytes) != 2*hashLen {
		fmt.Println("Incorrect Obj Length! ", oBytes)
		t.FailNow()
	}

	err := o2.FromBytes(oBytes)
	if err != nil {
		fmt.Println("Error while decoding obj: ", err)
		t.FailNow()
	}

	if string(o2.HashList[0].GetBytes()) != string(o.HashList[0].GetBytes()) || string(o2.HashList[1].GetBytes()) != string(o.HashList[1].GetBytes()) {
		fmt.Println("Incorrect decoding of obj: ", o2)
		t.Fail()
	}
}

func TestPubkey(t *testing.T) {
	p := new(EncryptedPubkey)
	var err error

	address := make([]byte, 25, 25)
	pubkey := [65]byte{'a'}
	p.AddrHash = MakeHash(address)
	p.IV, p.Payload, err = encryption.SymmetricEncrypt(address, string(pubkey[:]))
	if err != nil {
		fmt.Println("Could not encrypt pubkey: ", err)
		t.FailNow()
	}

	pBytes := p.GetBytes()
	if len(pBytes) != 144 {
		fmt.Println("Incorrect length for pubkey: ", pBytes)
		t.FailNow()
	}

	pubkey2 := new(EncryptedPubkey)
	err = pubkey2.FromBytes(pBytes)
	if err != nil {
		fmt.Println("Error decoding pubkey: ", err)
		t.Fail()
	}
	if string(pubkey2.AddrHash.GetBytes()) != string(p.AddrHash.GetBytes()) {
		fmt.Println("Incorrect Address Hash: ", pubkey2.AddrHash)
		t.FailNow()
	}

	pubkeyTest := encryption.SymmetricDecrypt(pubkey2.IV, address, pubkey2.Payload)
	if string(pubkeyTest[:65]) != string(pubkey[:]) {
		fmt.Println("Incorrect public key decryption: ", pubkeyTest)
		t.Fail()
	}
}

func TestPurge(t *testing.T) {
	p := new(Purge)
	p.Txid = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	pBytes := p.GetBytes()
	if len(pBytes) != 16 {
		fmt.Println("Error encoding purge: ", pBytes)
		t.FailNow()
	}

	p2 := new(Purge)
	p2.FromBytes(pBytes)
	if string(p2.Txid[:]) != string(p.Txid[:]) {
		fmt.Println("Incorrect decoding: ", p2.Txid)
		t.Fail()
	}
}

func TestMessage(t *testing.T) {
	log := make(chan string, 100)
	priv, x, y := encryption.CreateKey(log)
	pub := elliptic.Marshal(elliptic.P256(), x, y)
	address, _ := encryption.GetAddress(log, x, y)

	msg := new(Message)
	msg.AddrHash = MakeHash(address)
	msg.TxidHash = MakeHash([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	msg.Timestamp = time.Now().Round(time.Second)
	msg.Content = *encryption.Encrypt(log, pub, "Hello World!")

	mBytes := msg.GetBytes()
	if mBytes == nil {
		fmt.Println("Error Encoding Message!")
		t.FailNow()
	}

	msg2 := new(Message)
	msg2.FromBytes(mBytes)
	if string(msg2.AddrHash.GetBytes()) != string(msg.AddrHash.GetBytes()) || string(msg2.TxidHash.GetBytes()) != string(msg.TxidHash.GetBytes()) || msg2.Timestamp.Unix() != msg.Timestamp.Unix() {
		fmt.Println("Message Header incorrect: ", msg2)
		t.FailNow()
	}

	if string(encryption.Decrypt(log, priv, &msg.Content)[:12]) != "Hello World!" {
		fmt.Println("Message content incorrect: ", string(encryption.Decrypt(log, priv, &msg.Content)[:12]))
		t.Fail()
	}
}
