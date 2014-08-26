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
	"time"
)

const (
	LOCAL_VERSION = 1
	LOCAL_USER    = "emp v0.1"
	verLen        = 28
)

type Version struct {
	Version   uint16    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	IpAddress net.IP    `json:"ip_address"`
	Port      uint16    `json:"port"`
	UserAgent string    `json:"user_agent"`
}

func (v *Version) FromBytes(data []byte) error {
	if len(data) < verLen {
		return errors.New("Data too short!")
	}
	if v == nil {
		return errors.New("Could not load nil version.")
	}

	buffer := bytes.NewBuffer(data)

	v.Version = binary.BigEndian.Uint16(buffer.Next(2))
	v.Timestamp = time.Unix(int64(binary.BigEndian.Uint64(buffer.Next(8))), 0)
	v.IpAddress = net.IP(buffer.Next(16))
	v.Port = binary.BigEndian.Uint16(buffer.Next(2))
	v.UserAgent = buffer.String()
	return nil
}

func (v *Version) GetBytes() []byte {
	if v == nil {
		return nil
	}

	ret := make([]byte, verLen, verLen)

	binary.BigEndian.PutUint16(ret[:2], v.Version)
	binary.BigEndian.PutUint64(ret[2:10], uint64(v.Timestamp.Unix()))
	copy(ret[10:26], []byte(v.IpAddress))
	binary.BigEndian.PutUint16(ret[26:28], v.Port)
	ret = append(ret, v.UserAgent...)

	return ret
}
