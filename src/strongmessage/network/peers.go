package network

import (
	"fmt"
	"net"
	"time"
)

type Peer struct {
	IpAddress net.IP    `json:"ip_address"`
	Port      uint16    `json:"port"`
	LastSeen  time.Time `json:"last_seen"`
}

func (p *Peer) TcpString() string {
	return fmt.Sprintf("tcp://%s:%d", p.IpAddress.String(), p.Port)
}
