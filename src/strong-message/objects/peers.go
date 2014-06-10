package objects

import (
  "net"
  "time"
)

type Peer struct {
  IpAddress net.IP
  Port uint16
  LastSeen time.Time
}
