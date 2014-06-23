package objects

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

const (
	LOCAL_VERSION = 1
	LOCAL_USER    = "strongmsgd v0.1"
)

type Version struct {
	Version   uint32 `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	IpAddress net.IP `json:"ip_address"`
	Port      uint16 `json:"port"`
	AdminPort uint16 `json:"admin_port"`
	UserAgent string `json:"user_agent"`
}

func (v *Version) FromBytes(log chan string, data []byte) error {
	buffer := bytes.NewBuffer(data)
	enc := gob.NewDecoder(buffer)
	err := enc.Decode(v)
	if err != nil {
		log <- fmt.Sprintf("Version Decoding error: %s", err.Error())
		return err
	}
	return nil
}

func (v *Version) GetBytes(log chan string) []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(v)
	if err != nil {
		log <- "Encoding error!"
		log <- err.Error()
		return nil
	} else {
		return buffer.Bytes()
	}

}
