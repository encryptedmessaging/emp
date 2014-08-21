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
