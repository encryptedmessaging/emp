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

type MetaMessage struct {
	TxidHash  Hash      `json:"txid_hash"`
	Timestamp time.Time `json:"sent"`
	Purged    bool      `json:"read"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
}

type FullMessage struct {
	MetaMessage MetaMessage                  `json:"info"`
	Decrypted   *DecryptedMessage            `json:"decrypted"`
	Encrypted   *encryption.EncryptedMessage `json:"encrypted"`
}

type DecryptedMessage struct {
	Txid      [16]byte
	Pubkey    [65]byte
	Subject   string
	MimeType  string
	Length    uint32
	Content   string
	Signature [65]byte
}

func (d *DecryptedMessage) GetBytes() []byte {
	if d == nil {
		return nil
	}

	ret := append(d.Txid[:], d.Pubkey[:]...)
	ret = append(ret, d.Subject...)
	ret = append(ret, 0)
	ret = append(ret, d.MimeType...)
	ret = append(ret, 0)

	leng := make([]byte, 4, 4)

	binary.BigEndian.PutUint32(leng, d.Length)

	ret = append(ret, leng...)
	ret = append(ret, d.Content...)
	ret = append(ret, d.Signature[:]...)

	return ret
}

func (d *DecryptedMessage) FromBytes(data []byte) error {
	if d == nil {
		return errors.New("Can't fill nil object!")
	}

	var err error

	buf := bytes.NewBuffer(data)
	copy(d.Txid[:], buf.Next(16))
	copy(d.Pubkey[:], buf.Next(65))
	d.Subject, err = buf.ReadString(0)
	if err != nil {
		return err
	}
	d.Subject = d.Subject[:len(d.Subject)-1]
	d.MimeType, err = buf.ReadString(0)
	if err != nil {
		return err
	}
	d.MimeType = d.MimeType[:len(d.MimeType)-1]

	d.Length = binary.BigEndian.Uint32(buf.Next(4))

	d.Content = string(buf.Next(int(d.Length)))
	copy(d.Signature[:], buf.Next(65))

	return nil
}
