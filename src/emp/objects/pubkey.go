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
