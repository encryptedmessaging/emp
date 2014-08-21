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
	"crypto/sha512"
	"errors"
)

const (
	hashLen = 48
)

type Hash [hashLen]byte

func MakeHash(data []byte) Hash {
	hashArr := sha512.Sum384(data)
	return Hash(hashArr)
}

func (h *Hash) GetBytes() []byte {
	if h == nil {
		return nil
	}
	hashArr := [hashLen]byte(*h)
	return hashArr[:]
}

func (h *Hash) FromBytes(data []byte) error {
	if h == nil {
		return errors.New("Can't fill nil Hash Object.")
	}
	if len(data) != hashLen {
		return errors.New("Invalid hash length.")
	}
	for i := 0; i < hashLen; i++ {
		(*h)[i] = data[i]
	}

	return nil
}

type Obj struct {
	HashList []Hash
}

func (o *Obj) GetBytes() []byte {
	if o == nil {
		return nil
	}
	if o.HashList == nil {
		return nil
	}

	ret := make([]byte, 0, hashLen*len(o.HashList))
	for _, hash := range o.HashList {
		ret = append(ret, hash.GetBytes()...)
	}
	return ret
}

func (o *Obj) FromBytes(data []byte) error {
	if o == nil {
		return errors.New("Can't fill nil Obj Object.")
	}
	if len(data)%hashLen != 0 {
		return errors.New("Invalid hashlist Length!")
	}

	for i := 0; i < len(data); i += hashLen {
		h := new(Hash)
		err := h.FromBytes(data[i : i+hashLen])
		if err != nil {
			return err
		}
		o.HashList = append(o.HashList, *h)
	}
	return nil
}
