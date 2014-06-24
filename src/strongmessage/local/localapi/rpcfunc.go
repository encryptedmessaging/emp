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
			reply.Pubkey = getPubkey(service.Log, service.Config, reply.Hash, reply.Address)
		}
		return nil
	}

	return err
}



func (service *StrongService) AddUpdateAddress(r *http.Request, args *ShortAddress, reply *NilParam) error {
	halfaddr, err := base64.StdEncoding.DecodeString(args.Address[1:])
	if err != nil {
		return err
	}
	address := make([]byte, 1, 1)
	address[0] = 0x01
	address = append(address, halfaddr...)

	hashArr := sha512.Sum384(address)
	hash := hashArr[:]

	// Check Address
	if len(address) != 25 {
		return errors.New("Invalid Address: Incorrect Length")
	}

	sum := sha512.Sum384(address[:21])
	sum = sha512.Sum384(sum[:])

	if string(sum[:4]) != string(address[21:]) {
		return errors.New("Invalid Address: Bad Checksum")
	}

	// Check that pubkey matches address
	if args.Pubkey != nil {
		x, y := elliptic.Unmarshal(elliptic.P256(), args.Pubkey)
		if x == nil {
			return errors.New("Public Key Invalid")
		}
		address2, _ := encryption.GetAddress(service.Log, x, y)
		if string(address) != string(address2) {
			return errors.New("Public Key doesn't match provided address!")
		}
	}

	// Record Public Key for Network
	IV, pubkeyCipher, _ := encryption.SymmetricEncrypt(address, string(args.Pubkey))
	service.Config.RecvChan <- *network.NewFrame("pubkey", append(hash, append(IV[:], pubkeyCipher...)...))


	hashType := localdb.Contains(string(hash))
	switch hashType {
	case localdb.NOTFOUND:
		// Insert new address
		err := localdb.LocalDB.Exec("INSERT INTO addressbook VALUES (?, ?, ?, ?, ?)", hash, address, args.IsRegistered, args.Pubkey, args.Privkey)
		if err != nil {
			service.Log <- fmt.Sprintf("Error inserting address into localdb... %s", err)
			return err
		}
		localdb.Add(string(hash), localdb.ADDRESS)
		if args.Pubkey == nil {
			args.Pubkey = getPubkey(service.Log, service.Config, hash, address)
		}
		return nil
	case localdb.ADDRESS:
		// Update Address in database
		err := localdb.LocalDB.Exec("UPDATE addressbook SET registered=? WHERE hash=?", args.IsRegistered, hash)
		if err != nil {
			service.Log <- fmt.Sprintf("Error updating address in localdb... %s", err)
			return err
		}

		if args.Pubkey != nil {
			err = localdb.LocalDB.Exec("UPDATE addressbook SET pubkey=? WHERE hash=?", args.Pubkey, hash)
			if err != nil {
				service.Log <- fmt.Sprintf("Error updating pubkey in localdb... %s", err)
				return err
			}
		}

		if args.Privkey != nil {
			err = localdb.LocalDB.Exec("UPDATE addressbook SET privkey=? WHERE hash=?", args.Privkey, hash)
			if err != nil {
				service.Log <- fmt.Sprintf("Error updating privkey in localdb... %s", err)
				return err
			}
		}

		return nil
	default:
		return errors.New("Hash appears to be a Message TXID...")
	}
}

func (service *StrongService) ListAddresses(r *http.Request, args *bool, reply *([]ShortAddress)) error {

	for s, err := localdb.LocalDB.Query("SELECT address, registered, pubkey, privkey FROM addressbook"); err == nil; err = s.Next() {
		sa := new(ShortAddress)
		addr := make([]byte, 25, 25)
		s.Scan(&addr, &sa.IsRegistered, &sa.Pubkey, &sa.Privkey) // Assigns 1st column to rowid, the rest to row
		sa.Address = "1" + base64.StdEncoding.EncodeToString(addr[1:])
		if args != nil {
			if *args != sa.IsRegistered {
				continue
			}
		}
		*reply = append(*reply, *sa)
	}

	return nil
}