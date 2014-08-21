/**
    Copyright 2014 JARST, LLC
    
    This file is part of EMP.

    EMP is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
**/

package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"emp/encryption"
	"time"
)

type Message struct {
	AddrHash  Hash
	TxidHash  Hash
	Timestamp time.Time
	Content   encryption.EncryptedMessage
}

const (
	msgLen = 2*hashLen + 8
)

// Message Commands
const (
	VERSION = iota
	PEER    = iota
	OBJ     = iota
	GETOBJ  = iota

	PUBKEY_REQUEST = iota
	PUBKEY         = iota
	MSG            = iota
	PURGE          = iota

	CHECKTXID      = iota
	PUB            = iota
)

// Message Types
const (
	BROADCAST = iota
	REQUEST   = iota
	REPLY     = iota
)

func (m *Message) FromBytes(data []byte) error {
	if len(data) < msgLen {
		return errors.New("Data too short to create message!")
	}
	if m == nil {
		return errors.New("Can't fill nil Message object!")
	}
	buffer := bytes.NewBuffer(data)
	m.AddrHash.FromBytes(buffer.Next(hashLen))
	m.TxidHash.FromBytes(buffer.Next(hashLen))
	m.Timestamp = time.Unix(int64(binary.BigEndian.Uint64(buffer.Next(8))), 0)
	m.Content.FromBytes(buffer.Bytes())

	return nil

}

func (m *Message) GetBytes() []byte {
	if m == nil {
		return nil
	}

	ret := make([]byte, 0, msgLen)
	ret = append(m.AddrHash.GetBytes(), m.TxidHash.GetBytes()...)
	time := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(time, uint64(m.Timestamp.Unix()))
	ret = append(ret, time...)
	ret = append(ret, m.Content.GetBytes()...)
	return ret
}
