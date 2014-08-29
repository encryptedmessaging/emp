/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package localdb

import (
	"emp/encryption"
	"emp/objects"
	"errors"
	"time"
)

func AddUpdateAddress(address *objects.AddressDetail) error {
	localMutex.Lock()
	defer localMutex.Unlock()

	var err error

	if address.Address == nil {
		address.Address = encryption.StringToAddress(address.String)
	}

	if address.Address == nil {
		return errors.New("Invalid Address!")
	}

	addrHash := objects.MakeHash(address.Address)

	if Contains(addrHash) == ADDRESS { // Exists in message database, update pubkey, privkey, and registration
		err = LocalDB.Exec("UPDATE addressbook SET registered=?, subscribed=?, label=? WHERE hash=?", address.IsRegistered, address.IsSubscribed, address.Label, addrHash.GetBytes())
		if err != nil {
			return err
		}

		if address.Pubkey != nil {
			err = LocalDB.Exec("UPDATE addressbook SET pubkey=? WHERE hash=?", address.Pubkey, addrHash.GetBytes())
			if err != nil {
				return err
			}
		}

		if address.Privkey != nil {
			err = LocalDB.Exec("UPDATE addressbook SET privkey=? WHERE hash=?", address.Privkey, addrHash.GetBytes())
			if err != nil {
				return err
			}
		}

		if address.EncPrivkey != nil {
			err = LocalDB.Exec("UPDATE addressbook SET encprivkey=? WHERE hash=?", address.EncPrivkey, addrHash.GetBytes())
			if err != nil {
				return err
			}
		}

	} else { // Doesn't exist yet, insert it!
		err = LocalDB.Exec("INSERT INTO addressbook VALUES (?, ?, ?, ?, ?, ?, ?, ?)", addrHash.GetBytes(), address.Address, address.IsRegistered, address.Pubkey, address.Privkey, address.Label, address.IsSubscribed, address.EncPrivkey)
		if err != nil {
			return err
		}
		Add(addrHash, ADDRESS)
	}

	return nil
}

func GetAddressDetail(addrHash objects.Hash) (*objects.AddressDetail, error) {
	localMutex.Lock()
	defer localMutex.Unlock()

	if Contains(addrHash) != ADDRESS {
		return nil, errors.New("Address not found!")
	}

	ret := new(objects.AddressDetail)

	s, err := LocalDB.Query("SELECT address, registered, pubkey, privkey, label, subscribed, encprivkey FROM addressbook WHERE hash=?", addrHash.GetBytes())
	if err == nil {
		s.Scan(&ret.Address, &ret.IsRegistered, &ret.Pubkey, &ret.Privkey, &ret.Label, &ret.IsSubscribed, &ret.EncPrivkey)
		ret.String = encryption.AddressToString(ret.Address)
		return ret, nil
	}

	return nil, err
}

func ListAddresses(registered bool) [][2]string {
	ret := make([][2]string, 0, 0)

	for s, err := LocalDB.Query("SELECT address, label FROM addressbook WHERE registered=?", registered); err == nil; err = s.Next() {
		var addr []byte
		var label string
		s.Scan(&addr, &label)
		ret = append(ret, [2]string{encryption.AddressToString(addr), label})
	}

	return ret
}

func GetMessageDetail(txidHash objects.Hash) (*objects.FullMessage, error) {
	localMutex.Lock()
	defer localMutex.Unlock()

	if Contains(txidHash) > SENDBOX {
		return nil, errors.New("Message not found!")
	}

	ret := new(objects.FullMessage)
	ret.Encrypted = new(encryption.EncryptedMessage)
	ret.Decrypted = new(objects.DecryptedMessage)

	s, err := LocalDB.Query("SELECT * FROM msg WHERE txid_hash=?", txidHash.GetBytes())
	if err == nil {
		recipient := make([]byte, 0, 0)
		sender := make([]byte, 0, 0)
		encrypted := make([]byte, 0, 0)
		decrypted := make([]byte, 0, 0)
		txidHash := make([]byte, 0, 0)
		var timestamp int64
		var purged bool
		var box int

		s.Scan(&txidHash, &recipient, &timestamp, &box, &encrypted, &decrypted, &purged, &sender)
		ret.MetaMessage.TxidHash.FromBytes(txidHash)
		ret.MetaMessage.Recipient = encryption.AddressToString(recipient)
		ret.MetaMessage.Sender = encryption.AddressToString(sender)
		ret.MetaMessage.Timestamp = time.Unix(timestamp, 0)
		ret.MetaMessage.Purged = purged
		ret.Encrypted.FromBytes(encrypted)
		if len(decrypted) > 0 {
			ret.Decrypted.FromBytes(decrypted)
		} else {
			ret.Decrypted = nil
		}
		return ret, nil
	}

	return nil, err

}

func DeleteMessage(txidHash *objects.Hash) error {
	localMutex.Lock()
	defer localMutex.Unlock()

	if Contains(*txidHash) > SENDBOX {
		return errors.New("Error Deleting Message: Not Found!")
	}

	return LocalDB.Exec("DELETE FROM msg WHERE txid_hash=?", txidHash.GetBytes());
}

func DeleteAddress(addrHash *objects.Hash) error {
	localMutex.Lock()
	defer localMutex.Unlock()

	if Contains(*addrHash) > ADDRESS {
		return errors.New("Error Deleting Message: Not Found!")
	}

	return LocalDB.Exec("DELETE FROM addressbook WHERE hash=?", addrHash.GetBytes());
}

func AddUpdateMessage(msg *objects.FullMessage, box int) error {
	localMutex.Lock()
	defer localMutex.Unlock()

	var err error

	if Contains(msg.MetaMessage.TxidHash) > SENDBOX { // Insert Message Into Database!

		err = LocalDB.Exec("INSERT INTO msg VALUES (?, ?, ?, ?, ?, ?, ?, ?)", msg.MetaMessage.TxidHash.GetBytes(), encryption.StringToAddress(msg.MetaMessage.Recipient),
			msg.MetaMessage.Timestamp.Unix(), box, msg.Encrypted.GetBytes(), msg.Decrypted.GetBytes(), msg.MetaMessage.Purged, encryption.StringToAddress(msg.MetaMessage.Sender))
		if err != nil {
			return err
		}

	} else { // Update recipient, sender, purged, encrypted, decrypted, box
		if box < 0 {
			err = LocalDB.Exec("UPDATE msg SET purged=? WHERE txid_hash=?", msg.MetaMessage.Purged, msg.MetaMessage.TxidHash.GetBytes())
			if err != nil {
				return err
			}
		} else {
			err = LocalDB.Exec("UPDATE msg SET box=?, purged=? WHERE txid_hash=?", box, msg.MetaMessage.Purged, msg.MetaMessage.TxidHash.GetBytes())
			if err != nil {
				return err
			}
		}

		if len(msg.MetaMessage.Sender) > 0 {
			err = LocalDB.Exec("UPDATE msg SET sender=? WHERE txid_hash=?", encryption.StringToAddress(msg.MetaMessage.Sender), msg.MetaMessage.TxidHash.GetBytes())
			if err != nil {
				return err
			}
		}

		if len(msg.MetaMessage.Recipient) > 0 {
			err = LocalDB.Exec("UPDATE msg SET recipient=? WHERE txid_hash=?", encryption.StringToAddress(msg.MetaMessage.Recipient), msg.MetaMessage.TxidHash.GetBytes())
			if err != nil {
				return err
			}
		}

		if msg.Encrypted != nil {
			err = LocalDB.Exec("UPDATE msg SET encrypted=? WHERE txid_hash=?", msg.Encrypted.GetBytes(), msg.MetaMessage.TxidHash.GetBytes())
			if err != nil {
				return err
			}
		}

		if msg.Decrypted != nil {
			err = LocalDB.Exec("UPDATE msg SET decrypted=? WHERE txid_hash=?", msg.Decrypted.GetBytes(), msg.MetaMessage.TxidHash.GetBytes())
			if err != nil {
				return err
			}
		}

	}

	Add(msg.MetaMessage.TxidHash, box)
	return nil
}

func GetBox(box int) []objects.MetaMessage {
	if box > SENDBOX || box < INBOX {
		return nil
	}

	localMutex.Lock()
	defer localMutex.Unlock()

	ret := make([]objects.MetaMessage, 0, 0)

	for s, err := LocalDB.Query("SELECT txid_hash, timestamp, purged, sender, recipient FROM msg WHERE box=?", box); err == nil; err = s.Next() {
		mm := new(objects.MetaMessage)
		sendBytes := make([]byte, 0, 0)
		recvBytes := make([]byte, 0, 0)
		txidHash := make([]byte, 0, 0)
		var timestamp int64

		s.Scan(&txidHash, &timestamp, &mm.Purged, &sendBytes, &recvBytes)
		mm.Sender = encryption.AddressToString(sendBytes)
		mm.Recipient = encryption.AddressToString(recvBytes)

		mm.TxidHash.FromBytes(txidHash)
		mm.Timestamp = time.Unix(timestamp, 0)

		ret = append(ret, *mm)
	}

	return ret
}

func GetBySender(sender string) []objects.MetaMessage {
	localMutex.Lock()
	defer localMutex.Unlock()

	ret := make([]objects.MetaMessage, 0, 0)

	for s, err := LocalDB.Query("SELECT txid_hash, timestamp, purged, sender, recipient FROM msg WHERE sender=?", sender); err == nil; err = s.Next() {
		mm := new(objects.MetaMessage)
		sendBytes := make([]byte, 0, 0)
		recvBytes := make([]byte, 0, 0)
		txidHash := make([]byte, 0, 0)
		var timestamp int64

		s.Scan(&txidHash, &timestamp, &mm.Purged, &sendBytes, &recvBytes)
		mm.Sender = encryption.AddressToString(sendBytes)
		mm.Recipient = encryption.AddressToString(recvBytes)

		mm.TxidHash.FromBytes(txidHash)
		mm.Timestamp = time.Unix(timestamp, 0)

		ret = append(ret, *mm)
	}

	return ret
}

func GetByRecipient(recipient string) []objects.MetaMessage {
	localMutex.Lock()
	defer localMutex.Unlock()

	ret := make([]objects.MetaMessage, 0, 0)

	for s, err := LocalDB.Query("SELECT txid_hash, timestamp, purged, sender, recipient FROM msg WHERE recipient=?", recipient); err == nil; err = s.Next() {
		mm := new(objects.MetaMessage)
		sendBytes := make([]byte, 0, 0)
		recvBytes := make([]byte, 0, 0)
		txidHash := make([]byte, 0, 0)
		var timestamp int64

		s.Scan(&txidHash, &timestamp, &mm.Purged, &sendBytes, &recvBytes)
		mm.Sender = encryption.AddressToString(sendBytes)
		mm.Recipient = encryption.AddressToString(recvBytes)

		mm.TxidHash.FromBytes(txidHash)
		mm.Timestamp = time.Unix(timestamp, 0)

		ret = append(ret, *mm)
	}

	return ret
}

func DeleteObject(obj objects.Hash) error {
	var err error
	switch Contains(obj) {
	case INBOX:
		fallthrough
	case SENDBOX:
		fallthrough
	case OUTBOX:
		err = LocalDB.Exec("DELETE FROM msg WHERE txid_hash=?", obj.GetBytes())
	case ADDRESS:
		err = LocalDB.Exec("DELETE FROM addressbook WHERE hash=?", obj.GetBytes())
	default:
		err = errors.New("Hash not found!")
	}

	if err == nil {
		Del(obj)
	}

	return err
}
