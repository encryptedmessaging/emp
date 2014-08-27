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
	"encoding/binary"
	"errors"
	"emp/encryption"
	"time"
)

type MetaMessage struct {
	TxidHash  Hash      `json:"txid_hash"` // Hash of random identifier
	Timestamp time.Time `json:"sent"`      // Time message was sent
	Purged    bool      `json:"read"`      // Whether purge token has been received
	Sender    string    `json:"sender"`    // String representation of sender's address, if available.
	Recipient string    `json:"recipient"` // String representation of recipient's address, if available.
}

type FullMessage struct {
	MetaMessage MetaMessage                  `json:"info"`
	Decrypted   *DecryptedMessage            `json:"decrypted"`
	Encrypted   *encryption.EncryptedMessage `json:"encrypted"`
}

type DecryptedMessage struct {
	Txid      [16]byte // Randomly generated identifier and purge token.
	Pubkey    [65]byte // Sender's 65-byte Public Key
	Subject   string   // Human-readable subject of this message
	MimeType  string   // Mime-Type of Content
	Length    uint32   // Length of Content in bytes
	Content   string   // Content of message, could be any data.
	Signature [65]byte // Sender's Signature of entire message.
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
