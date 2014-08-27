/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package encryption

import "errors"

type EncryptedMessage struct {
	IV         [16]byte // Initialization Vector for AES encryption
	PublicKey  [65]byte // Random Public Key used for decryption
	CipherText []byte   // CipherText, length is multiple of AES blocksize
	HMAC       [32]byte // HMAC-SHA256, used to validate key before decryption
}

const (
	ivLen     = 16
	pubkeyLen = 65
	hmacLen   = 32
	minLen    = ivLen + pubkeyLen + hmacLen
)

func (ret *EncryptedMessage) FromBytes(b []byte) error {
	if len(b) < minLen {
		return errors.New("Bytes too short to create EncryptedMessage object.")
	}
	if ret == nil {
		return errors.New("Can't fill nil object.")
	}

	copy(ret.IV[:], b[:ivLen])
	copy(ret.PublicKey[:], b[ivLen:ivLen+pubkeyLen])
	ret.CipherText = append(ret.CipherText, b[ivLen+pubkeyLen:len(b)-hmacLen]...)
	copy(ret.HMAC[:], b[len(b)-hmacLen:])

	return nil
}

func (e *EncryptedMessage) GetBytes() []byte {
	if e == nil {
		return nil
	}
	ret := make([]byte, 0, 0)
	ret = append(ret, e.IV[:]...)
	ret = append(ret, e.PublicKey[:]...)
	ret = append(ret, e.CipherText...)
	ret = append(ret, e.HMAC[:]...)

	return ret
}
