package objects

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strongmessage/encryption"
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
	Encrypted   *encryption.EncryptedMessage `json:"encrypted"`
	Decrypted   *DecryptedMessage            `json:"decrypted"`
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
	d.MimeType, err = buf.ReadString(0)
	if err != nil {
		return err
	}

	d.Length = binary.BigEndian.Uint32(buf.Next(4))

	d.Content = string(buf.Next(int(d.Length)))
	copy(d.Signature[:], buf.Next(65))

	return nil
}
