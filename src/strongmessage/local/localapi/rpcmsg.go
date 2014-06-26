package localapi

import (
	"errors"
	"crypto/sha512"
	"strongmessage/local/localdb"
	"strongmessage/objects"
	"strongmessage/network"
	"strongmessage/encryption"
	"crypto/ecdsa"
	"crypto/elliptic"
	"net/http"
	"crypto/rand"
	"math/big"
	"time"
	"fmt"
	"encoding/base64"
)

type SendMsg struct {
	Sender string `json:"sender"`
	Recipient string `json:"recipient"`
	Plaintext string `json:"content"`
	Signature []byte `json:"signature"`
}

type SendResponse struct {
	TxidHash []byte `json:"txid_hash"`
	IsSent bool `json:"sent"`
}

func (service *StrongService) SendMessage(r *http.Request, args *SendMsg, reply *SendResponse) error {
	// Nil Check
	if len(args.Sender) == 0 || len(args.Recipient) == 0 || len(args.Plaintext) == 0 {
		return errors.New("All fields required except signature.")
	}

	// Get Addresses
	halfsender, err := base64.StdEncoding.DecodeString(args.Sender[1:])
	if err != nil {
		return err
	}
	sender := make([]byte, 1, 1)
	sender[0] = 0x01
	sender = append(sender, halfsender...)

	senderArr := sha512.Sum384(sender)
	senderHash := senderArr[:]

	halfrecipient, err := base64.StdEncoding.DecodeString(args.Recipient[1:])
	if err != nil {
		return err
	}
	recipient := make([]byte, 1, 1)
	recipient[0] = 0x01
	recipient = append(recipient, halfrecipient...)

	recipientArr := sha512.Sum384(recipient)
	recipientHash := recipientArr[:]

	// Check Addresses
	if len(sender) != 25 {
		return errors.New("Invalid Sender Address: Incorrect Length")
	}

	sum := sha512.Sum384(sender[:21])
	sum = sha512.Sum384(sum[:])

	if string(sum[:4]) != string(sender[21:]) {
		return errors.New("Invalid Sender Address: Bad Checksum")
	}

	if len(recipient) != 25 {
		return errors.New("Invalid Recipient Address: Incorrect Length")
	}

	sum = sha512.Sum384(recipient[:21])
	sum = sha512.Sum384(sum[:])

	if string(sum[:4]) != string(recipient[21:]) {
		return errors.New("Invalid Recipient Address: Bad Checksum")
	}

	// Ensure addresses are in database
	if localdb.Contains(string(senderHash)) != localdb.ADDRESS {
		return errors.New("Sender Address not in local database! Call AddUpdateAddress() First!")
	}
	if localdb.Contains(string(recipientHash)) != localdb.ADDRESS {
		return errors.New("Recipient Address not in local database! Call AddUpdateAddress() First!")
	}

	// Create New Message
	msg := new(LocalMessage)

	txid := make([]byte, 16, 16)

	n, err := rand.Reader.Read(txid)
	if err != nil || n != 16 {
		return err
	}
	txidArr := sha512.Sum384(txid)
	msg.TxidHash = txidArr[:]
	msg.AddrHash = recipientHash
	msg.Timestamp = time.Now().Round(time.Second)
	msg.IsPurged = false

	msg.Decrypted = new(objects.MessageUnencrypted)
	msg.Decrypted.Signature = nil
	msg.Decrypted.Txid = txid
	msg.Decrypted.SendAddr = sender
	msg.Decrypted.Timestamp = msg.Timestamp
	msg.Decrypted.DataType = "text/plain"
	msg.Decrypted.Data = append(msg.Decrypted.Data, args.Plaintext...)


	// Get Signature
	if args.Signature != nil {
		if len(args.Signature) != 65 {
			return errors.New("Invalid signature, should be valid eliptic public key!")
		}
		msg.Decrypted.Signature = args.Signature
	} else {
		for s, err := localdb.LocalDB.Query("SELECT privkey FROM addressbook"); err == nil; err = s.Next() {
			privBytes := make([]byte, 0, 64)
			s.Scan(&privBytes)
			if len(privBytes) > 0 {
				priv := new(ecdsa.PrivateKey)
				priv.D = new(big.Int).SetBytes(privBytes)
				priv.PublicKey.Curve = elliptic.P256()
				r, s, err := ecdsa.Sign(rand.Reader, priv, msg.Decrypted.Data)
				if err != nil {
					return err
				}
				msg.Decrypted.Signature = elliptic.Marshal(elliptic.P256(), r, s)
			}
		}
	}

	if msg.Decrypted.Signature == nil {
		return errors.New("Could not sign message: no private key for send address. Please provide key or signature!")
	}

	pubkey := getPubkey(service.Log, service.Config, msg.AddrHash, recipient)

	if len(pubkey) == 0 {
		reply.IsSent = false
		// Add message to outbox...
		err = localdb.LocalDB.Exec("INSERT INTO msg VALUES (?, NULL, ?, ?)", msg.TxidHash, msg.Decrypted.GetBytes(), msg.IsPurged)
		if err != nil {
			service.Log <- fmt.Sprintf("Error inserting message into localdb... %s", err)
			return err
		}

		err = localdb.LocalDB.Exec("INSERT INTO outbox VALUES (?, ?, ?, ?)", msg.TxidHash, msg.Timestamp.Unix(), sender, recipient)
		if err != nil {
			service.Log <- fmt.Sprintf("Error inserting message into outbox in localdb... %s", err)
			return err
		}

		localdb.Add(string(msg.TxidHash), localdb.OUTBOX)

	} else {
		// Send message and add to sendbox...

		msg.Encrypted = encryption.Encrypt(service.Log, pubkey, string(msg.Decrypted.Data))

		err = localdb.LocalDB.Exec("INSERT INTO msg VALUES (?, ?, ?, ?)", msg.TxidHash, msg.Encrypted.GetBytes(), msg.Decrypted.GetBytes(), msg.IsPurged)
		if err != nil {
			service.Log <- fmt.Sprintf("Error inserting message into localdb... %s", err)
			return err
		}

		err = localdb.LocalDB.Exec("INSERT INTO sendbox VALUES (?, ?, ?, ?)", msg.TxidHash, msg.Timestamp.Unix(), sender, recipient)
		if err != nil {
			service.Log <- fmt.Sprintf("Error inserting message into outbox in localdb... %s", err)
			return err
		}


		sendMsg := new(objects.Message)
		sendMsg.AddrHash = msg.AddrHash
		sendMsg.TxidHash = msg.TxidHash
		sendMsg.Timestamp = msg.Timestamp
		sendMsg.Content = *msg.Encrypted

		service.Config.RecvChan <- *network.NewFrame("msg", sendMsg.GetBytes(service.Log))

		localdb.Add(string(msg.TxidHash), localdb.SENDBOX)
		reply.IsSent = true
	}

	reply.TxidHash = msg.TxidHash
	return nil
}

func (service *StrongService) Inbox(r *http.Request, args *NilParam, reply *[]MetaMessage) error {
	for s, err := localdb.LocalDB.Query("SELECT * FROM inbox"); err == nil; err = s.Next() {
		msg := new(MetaMessage)
		sender := make([]byte, 0, 25)
		recipient := make([]byte, 0, 25)
		var timestamp int64
		s.Scan(&msg.TxidHash, &timestamp, &sender, &recipient)
		msg.Timestamp = time.Unix(timestamp, 0)
		if len(sender) == 25 {
			msg.Sender = "1" + base64.StdEncoding.EncodeToString(sender[1:])
		}
		if len(recipient) == 25 {
			msg.Recipient = "1" + base64.StdEncoding.EncodeToString(recipient[1:])
		}
		*reply = append(*reply, *msg)
	}
	return nil
}

func (service *StrongService) Outbox(r *http.Request, args *NilParam, reply *[]MetaMessage) error {
	for s, err := localdb.LocalDB.Query("SELECT * FROM outbox"); err == nil; err = s.Next() {
		msg := new(MetaMessage)
		sender := make([]byte, 0, 25)
		recipient := make([]byte, 0, 25)
		var timestamp int64
		s.Scan(&msg.TxidHash, &timestamp, &sender, &recipient)
		msg.Timestamp = time.Unix(timestamp, 0)
		if len(sender) == 25 {
			msg.Sender = "1" + base64.StdEncoding.EncodeToString(sender[1:])
		}
		if len(recipient) == 25 {
			msg.Recipient = "1" + base64.StdEncoding.EncodeToString(recipient[1:])
		}
		*reply = append(*reply, *msg)
	}
	return nil
}

func (service *StrongService) Sendbox(r *http.Request, args *NilParam, reply *[]MetaMessage) error {
	for s, err := localdb.LocalDB.Query("SELECT * FROM sendbox"); err == nil; err = s.Next() {
		msg := new(MetaMessage)
		sender := make([]byte, 0, 25)
		recipient := make([]byte, 0, 25)
		var timestamp int64
		s.Scan(&msg.TxidHash, &timestamp, &sender, &recipient)
		msg.Timestamp = time.Unix(timestamp, 0)
		if len(sender) == 25 {
			msg.Sender = "1" + base64.StdEncoding.EncodeToString(sender[1:])
		}
		if len(recipient) == 25 {
			msg.Recipient = "1" + base64.StdEncoding.EncodeToString(recipient[1:])
		}
		*reply = append(*reply, *msg)
	}
	return nil
}

func (service *StrongService) OpenMessage(r *http.Request, args *[]byte, reply *LocalMessage) error {
	reply.TxidHash = *args
	if len(reply.TxidHash) != 48 {
		return errors.New("Invalid Txid for message!")
	}
	sql := "SELECT timestamp, recipient, encrypted, decrypted, purged FROM %s INNER JOIN msg ON %s.txid_hash=msg.txid_hash WHERE msg.txid_hash=?"
	var table string

	switch localdb.Contains(string(reply.TxidHash)) {
	case localdb.INBOX:
		table = "inbox"
	case localdb.SENDBOX:
		table = "sendbox"
	case localdb.OUTBOX:
		table = "outbox"
	default:
		return errors.New("Message not found!")
	}
	sql = fmt.Sprintf(sql, table, table)

	for s, err := localdb.LocalDB.Query(sql, reply.TxidHash); err == nil; err = s.Next() {
		recipient := make([]byte, 0, 25)
		enc := make([]byte, 0, 0)
		dec := make([]byte, 0, 0)
		var timestamp int64
		s.Scan(&timestamp, &recipient, &enc, &dec, &reply.IsPurged)
		reply.Timestamp = time.Unix(timestamp, 0)
		addrArr := sha512.Sum384(recipient)
		reply.AddrHash = addrArr[:]
		reply.Encrypted = objects.EncryptedFromBytes(enc)
		reply.Decrypted = objects.DecryptedFromBytes(dec)
		if reply.Decrypted == nil {
			// Attempt to decrypt message, purge if successful
			s2, err := localdb.LocalDB.Query("SELECT privkey FROM addressbook WHERE hash=?", reply.AddrHash)
			if err == nil {
				privkey := make([]byte, 0, 0)
				s2.Scan(&privkey)
				reply.Decrypted = objects.DecryptedFromBytes(encryption.Decrypt(service.Log, privkey, reply.Encrypted))
				if reply.Decrypted != nil {
					// Decryption Successful! Update Databases and Purge
					err = localdb.LocalDB.Exec("UPDATE msg SET decrypted=?, purged=? WHERE txid_hash=?", reply.Decrypted.GetBytes(), reply.IsPurged, reply.TxidHash)
					if err != nil {
						service.Log <- fmt.Sprintf("OpenMessage(): Error Updating msg Database... %s", err.Error())
						return err
					}
					sql2 := fmt.Sprintf("UPDATE %s SET sender=? WHERE txid_hash=?", table)
					err = localdb.LocalDB.Exec(sql2, reply.Decrypted.SendAddr, reply.TxidHash)
					if err != nil {
						service.Log <- fmt.Sprintf("OpenMessage(): Error Updating %s database... %s", table, err.Error())
						return err
					}
					service.Config.RecvChan <- *network.NewFrame("purge", reply.Decrypted.Txid)
					return nil
				}
			}
		}
		return nil
	}
	return errors.New("Message not found!")
}