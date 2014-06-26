package network

import (
	"encoding/binary"
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"net"
	"time"
)

const (
	PeerLen = 28
)

type Peer struct {
	IpAddress net.IP    `json:"ip_address"`
	Port      uint16    `json:"port"`
	AdminPort uint16    `json:"admin_port"`
	LastSeen  time.Time `json:"last_seen"`
	socket    *zmq.Socket
}

func (p *Peer) GetBytes() []byte {
	ret := make([]byte, PeerLen, PeerLen)
	copy(ret[0:16], []byte(p.IpAddress.To16()))

	binary.BigEndian.PutUint16(ret[16:18], p.Port)
	binary.BigEndian.PutUint16(ret[18:20], p.AdminPort)
	binary.BigEndian.PutUint64(ret[20:28], uint64(p.LastSeen.Unix()))

	return ret
}

func (p *Peer) FromBytes(b []byte) error {
	if p == nil {
		return errors.New("Can't fill nil Peer object...")
	}
	if len(b) < PeerLen {
		return errors.New("Byte slice too shor")
	}

	p.IpAddress = net.IP(append([]byte(p.IpAddress), b[:16]...))
	p.Port = binary.BigEndian.Uint16(b[16:18])
	p.AdminPort = binary.BigEndian.Uint16(b[18:20])
	p.LastSeen = time.Unix(int64(binary.BigEndian.Uint64(b[20:28])), 0)
	p.socket = nil
	return nil
}

func (p *Peer) TcpString() string {
	return fmt.Sprintf("tcp://%s:%d", p.IpAddress.String(), p.Port)
}

func (p *Peer) AdminTcpString() string {
	return fmt.Sprintf("tcp://%s:%d", p.IpAddress.String(), p.AdminPort)
}

func (p *Peer) Connect(log chan string, context *zmq.Context) error {
	if p.socket != nil {
		return nil
	}

	// Setup Socket
	socket, err := context.NewSocket(zmq.REQ)
	if err != nil {
		log <- "error creating socket"
		log <- err.Error()
		return err
	}

	// Connect to Socket
	err = socket.Connect(p.AdminTcpString())
	if err != nil {
		log <- "Error subscribing to peer..."
		log <- err.Error()
		socket.Close()
		return err
	}

	p.socket = socket
	return nil
}

func (p *Peer) Disconnect() {
	if p.socket == nil {
		return
	}

	p.socket.Close()
	p.socket = nil
}

func (p *Peer) SendRequest(log chan string, frame *Frame, recvChannel chan Frame) bool {
	if p.socket == nil {
		return false
	}


	go func() {
		err := p.socket.Send(frame.GetBytes(), 0)
		if err != nil {
			log <- "Error sending frame..."
			log <- err.Error()
			return
		}
		data, err := p.socket.Recv(0)

		if err != nil {
			log <- "Error receiving from socket..."
			log <- err.Error()
			return
		}

		frame, err = FrameFromBytes(data)
		if err != nil {
			log <- "Received invalid frame..."
			log <- err.Error()
		}

		frame.Peer = p

		recvChannel <- *frame
	}()
	return true
}
