package localapi

import (
	"strongmessage/objects"
	"time"
)

type LocalMessage struct {
	TxidHash  []byte
	AddrHash  []byte
	Timestamp time.Time
	Encrypted *objects.EncryptedData
	Decrypted *objects.MessageUnencrypted
}

type LocalAddress struct {
	Hash         []byte `json:"address_hash"`
	Address      []byte `json:"address_bytes"`
	String       string `json:"address"`
	IsRegistered bool   `json:"registered"`
	Pubkey       []byte `json:"pubkey"`
	Privkey      []byte `json:"privkey"`
}

type ShortAddress struct {
	Address string `json:"address"`
	IsRegistered bool `json:"registered"`
	Pubkey []byte `json:"pubkey"`
	Privkey []byte `json:"privkey"`
}