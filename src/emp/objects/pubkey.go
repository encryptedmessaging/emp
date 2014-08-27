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
	"errors"
)

const (
	encPubLen = 144
)

type EncryptedPubkey struct {
	AddrHash Hash     // Hash of address that own this public key.
	IV       [16]byte // IV for AES-256 encryption of public key
	Payload  []byte   // Public key encrypted with AES-256. The Address is the key.
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
