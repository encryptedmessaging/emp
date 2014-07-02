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
	IP       net.IP
	Port     uint16
	LastSeen time.Time
}

type NodeList struct {
	Nodes map[string]Node
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

	for i := 0; i < len(data); i += nodeLen {
		b := bytes.NewBuffer(data[i : i+nodeLen])
		node := new(Node)
		node.IP = net.IP(b.Next(16))
		node.Port = binary.BigEndian.Uint16(b.Next(2))
		node.LastSeen = time.Unix(int64(binary.BigEndian.Uint64(b.Next(8))), 0)
		n.Nodes[node.String()] = *node
	}

	return nil
}
