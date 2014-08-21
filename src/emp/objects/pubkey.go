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
	"errors"
)

const (
	encPubLen = 144
)

type EncryptedPubkey struct {
	AddrHash Hash
	IV       [16]byte
	Payload  []byte
}

func (e *EncryptedPubkey) GetBytes() []byte {
	if e == nil {
		return nil
	}

	ret := make([]byte, hashLen, encPubLen)

	copy(ret, e.AddrHash.GetBytes())
	ret = append(ret, e.IV[:]...)
	ret = append(ret, e.Payload[:]...)
	return ret
}

func (e *EncryptedPubkey) FromBytes(data []byte) error {
	if e == nil {
		return errors.New("Can't fill nil EncryptedPubkey Object.")
	}
	if len(data) < encPubLen {
		return errors.New("Data too short for encrypted public key.")
	}

	b := bytes.NewBuffer(data)
	e.AddrHash.FromBytes(b.Next(hashLen))
	copy(e.IV[:], b.Next(16))
	e.Payload = append(e.Payload, b.Next(80)...)
	return nil
}
