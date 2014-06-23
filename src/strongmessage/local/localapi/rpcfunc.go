package localapi

import (
	"crypto/elliptic"
	"crypto/sha512"
	"errors"
	"fmt"
	"net/http"
	"strongmessage/encryption"

	"encoding/base64"
	"strongmessage/local/localdb"
	"strongmessage/network"
)

var logChan chan string

func (s *StrongService) CreateAddress(r *http.Request, args *NilParam, reply *LocalAddress) error {

	// Create Address

	priv, x, y := encryption.CreateKey(s.Log)
	reply.Privkey = priv
	if x == nil {
		return errors.New("Key Pair Generation Error")
	}

	reply.Pubkey = elliptic.Marshal(elliptic.P256(), x, y)

	reply.IsRegistered = true

	reply.Address, reply.String = encryption.GetAddress(s.Log, x, y)

	if reply.Address == nil {
		return errors.New("Could not create address, function returned nil.")
	}

	sum := sha512.Sum384(reply.Address)
	reply.Hash = sum[:]

	// Add Address to Database
	err := localdb.LocalDB.Exec("INSERT INTO addressbook VALUES (?, ?, ?, ?, ?)", reply.Hash, reply.Address, reply.IsRegistered, reply.Pubkey, reply.Privkey)
	if err != nil {
		str := fmt.Sprintf("Error inserting new address into db... %s", err)
		s.Log <- str

		return errors.New(str)
	}

	localdb.Add(string(reply.Hash), localdb.ADDRESS)

	IV, pubkeyCipher, _ := encryption.SymmetricEncrypt(reply.Address, string(reply.Pubkey))

	// Record Pubkey for Network
	s.Config.RecvChan <- *network.NewFrame("pubkey", append(reply.Hash, append(IV[:], pubkeyCipher...)...))
	return nil
}

func (service *StrongService) GetAddress(r *http.Request, args *string, reply *LocalAddress) error {
	address, err := base64.StdEncoding.DecodeString((*args)[1:])
	if err != nil {
		return err
	}
	reply.Address = make([]byte, 1, 1)
	reply.Address[0] = 0x01
	reply.Address = append(reply.Address, address...)

	hashArr := sha512.Sum384(reply.Address)
	reply.Hash = hashArr[:]
	reply.String = *args

	if localdb.Contains(string(reply.Hash)) != localdb.ADDRESS {
		return errors.New(fmt.Sprintf("Address %s not found in local database! Use AddUpdateAddress() to fix.", reply.String))
	}

	for s, err := localdb.LocalDB.Query("SELECT registered, pubkey, privkey FROM addressbook WHERE hash=?", reply.Hash); err == nil; err = s.Next() {
		s.Scan(&reply.IsRegistered, &reply.Pubkey, &reply.Privkey) // Assigns 1st column to rowid, the rest to row
		if len(reply.Pubkey) == 0 {
			service.Config.RecvChan <- *network.NewFrame("pubkeyrq", reply.Hash)
		}
		return nil
	}

	return err
}
