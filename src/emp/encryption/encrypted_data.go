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

package encryption

import "errors"

type EncryptedMessage struct {
	IV         [16]byte
	PublicKey  [65]byte
	CipherText []byte
	HMAC       [32]byte
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
