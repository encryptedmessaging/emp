/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package objects

import "errors"

const (
	purgeLen = 16
)

type Purge struct {
	Txid [16]byte
}

func (p *Purge) GetBytes() []byte {
	if p == nil {
		return nil
	}

	ret := make([]byte, 0, purgeLen)

	ret = append(ret, p.Txid[:]...)

	return ret
}

func (p *Purge) FromBytes(data []byte) error {
	if p == nil {
		return errors.New("Can't fill nil Purge Object.")
	}
	if len(data) != purgeLen {
		return errors.New("Data too short for encrypted public key.")
	}

	copy(p.Txid[:], data)

	return nil
}
