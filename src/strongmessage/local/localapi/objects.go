package localapi

import (
	"strongmessage/objects"
	"time"
)

// Message Objects
type LocalMessage struct {
	TxidHash  []byte
	AddrHash  []byte
	Timestamp time.Time
	IsPurged  bool
	Encrypted *objects.EncryptedData
	Decrypted *objects.MessageUnencrypted
}


// Address Objects
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