/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"
)

const (
	nodeLen = 26
)

type Node struct {
	IP       net.IP    // Public IPv6 or IPv4 Address
	Port     uint16    // Port on which TCP Server is running
	LastSeen time.Time // Time of last connection to Node.
	Attempts uint8     // Number of reconnection attempt. Currently, node is forgotten after 3 failed attempts.
}

type NodeList struct {
	Nodes map[string]Node
}

func (n *Node) FromString(hostPort string) error {
	if n == nil {
		return errors.New("Can't fill nil object.")
	}

	ip, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil
	}

	n.IP = net.ParseIP(ip)
	prt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	n.Port = uint16(prt)
	n.LastSeen = time.Now().Round(time.Second)
	n.Attempts = 0
	return nil
}

func (n *Node) String() string {
	if n == nil {
		return ""
	}
	return net.JoinHostPort(n.IP.String(), strconv.Itoa(int(n.Port)))
}

func (n *NodeList) GetBytes() []byte {
	if n == nil {
		return nil
	}
	if n.Nodes == nil {
		return nil
	}

	ret := make([]byte, 0, nodeLen*len(n.Nodes))

	for _, node := range n.Nodes {
		nBytes := make([]byte, nodeLen, nodeLen)
		copy(nBytes, []byte(node.IP))
		binary.BigEndian.PutUint16(nBytes[16:18], node.Port)
		binary.BigEndian.PutUint64(nBytes[18:26], uint64(node.LastSeen.Unix()))
		ret = append(ret, nBytes...)
	}

	return ret
}

func (n *NodeList) FromBytes(data []byte) error {
	if len(data)%nodeLen != 0 {
		return errors.New("Incorrect length for a Node List.")
	}
	if n == nil {
		return errors.New("Can't configure nil Node List")
	}
	if n.Nodes == nil {
		n.Nodes = make(map[string]Node)
	}

	for i := 0; i < len(data); i += nodeLen {
		b := bytes.NewBuffer(data[i : i+nodeLen])
		node := new(Node)
		node.IP = net.IP(b.Next(16))
		node.Port = binary.BigEndian.Uint16(b.Next(2))
		node.LastSeen = time.Unix(int64(binary.BigEndian.Uint64(b.Next(8))), 0)
		node.Attempts = 0
		n.Nodes[node.String()] = *node
	}

	return nil
}
