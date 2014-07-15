package localapi

import (
	"crypto/ecdsa"
	"crypto/rand"
	"emp/encryption"
	"emp/local/localdb"
	"emp/objects"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type SendMsg struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Plaintext string `json:"content"`
}

type SendResponse struct {
	TxidHash []byte `json:"txid_hash"`
	IsSent   bool   `json:"sent"`
}

func (service *EMPService) SendMessage(r *http.Request, args *SendMsg, reply *SendResponse) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	// Nil Check
	if len(args.Sender) == 0 || len(args.Recipient) == 0 || len(args.Plaintext) == 0 {
		return errors.New("All fields required except signature.")
	}

	var err error

	// Get Addresses
	sendAddr := encryption.StringToAddress(args.Sender)
	if len(sendAddr) == 0 {
		return errors.New("Invalid sender address!")
	}

	recvAddr := encryption.StringToAddress(args.Recipient)
	if len(recvAddr) == 0 {
		return errors.New("Invalid recipient address!")
	}

	sender, err := localdb.GetAddressDetail(objects.MakeHash(sendAddr))
	if err != nil {
		return errors.New(fmt.Sprintf("Error pulling send address from Database: %s", err))
	}
	if sender.Pubkey == nil {
		sender.Pubkey = checkPubkey(service.Config, objects.MakeHash(sendAddr))
		if sender.Pubkey == nil {
			return errors.New("Sender's Public Key is required to send message!")
		}
	}
	if sender.Privkey == nil {
		return errors.New("SendMsg() requires a stored private key. Use SendRawMsg() instead.")
	}

	recipient, err := localdb.GetAddressDetail(objects.MakeHash(recvAddr))
	if err != nil {
		return errors.New(fmt.Sprintf("Error pulling recipient address from Database: %s", err))
	}

	// Create New Message
	msg := new(objects.FullMessage)
	msg.Decrypted = new(objects.DecryptedMessage)
	msg.Encrypted = nil

	// Fill out decrypted message
	n, err := rand.Read(msg.Decrypted.Txid[:])
	if n < len(msg.Decrypted.Txid[:]) || err != nil {
		return errors.New(fmt.Sprintf("Problem with random reader: %s", err))
	}
	copy(msg.Decrypted.Pubkey[:], sender.Pubkey)
	msg.Decrypted.Subject = args.Subject
	msg.Decrypted.MimeType = "text/plain"
	msg.Decrypted.Content = args.Plaintext
	msg.Decrypted.Length = uint32(len(msg.Decrypted.Content))

	// Fill Out Meta Message (save timestamp)
	msg.MetaMessage.Purged = false
	msg.MetaMessage.TxidHash = objects.MakeHash(msg.Decrypted.Txid[:])
	msg.MetaMessage.Sender = sender.String
	msg.MetaMessage.Recipient = recipient.String

	// Get Signature
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = encryption.GetCurve()
	priv.D = new(big.Int)
	priv.D.SetBytes(sender.Privkey)

	sign := msg.Decrypted.GetBytes()
	sign = sign[:len(sign)-65]
	signHash := objects.MakeHash(sign)

	x, y, err := ecdsa.Sign(rand.Reader, priv, signHash.GetBytes())
	if err != nil {
		return err
	}

	copy(msg.Decrypted.Signature[:], encryption.MarshalPubkey(x, y))

	// Check for pubkey
	if recipient.Pubkey == nil {
		recipient.Pubkey = checkPubkey(service.Config, objects.MakeHash(recipient.Address))
	}

	if recipient.Pubkey == nil {
		reply.IsSent = false
		// Add message to outbox...
		err = localdb.AddUpdateMessage(msg, localdb.OUTBOX)
		if err != nil {
			return err
		}

	} else {
		// Send message and add to sendbox...
		msg.Encrypted = encryption.Encrypt(service.Config.Log, recipient.Pubkey, string(msg.Decrypted.GetBytes()))
		msg.MetaMessage.Timestamp = time.Now().Round(time.Second)

		err = localdb.AddUpdateMessage(msg, localdb.SENDBOX)
		if err != nil {
			return err
		}

		sendMsg := new(objects.Message)
		sendMsg.TxidHash = msg.MetaMessage.TxidHash
		sendMsg.AddrHash = objects.MakeHash(recipient.Address)
		sendMsg.Timestamp = msg.MetaMessage.Timestamp
		sendMsg.Content = *msg.Encrypted

		service.Config.RecvQueue <- *objects.MakeFrame(objects.MSG, objects.BROADCAST, sendMsg)

		reply.IsSent = true
	}

	// Finish by setting msg's txid
	reply.TxidHash = msg.MetaMessage.TxidHash.GetBytes()
	return nil
}

func (service *EMPService) Inbox(r *http.Request, args *NilParam, reply *[]objects.MetaMessage) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	*reply = localdb.GetBox(localdb.INBOX)
	return nil
}

func (service *EMPService) Outbox(r *http.Request, args *NilParam, reply *[]objects.MetaMessage) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	*reply = localdb.GetBox(localdb.OUTBOX)
	return nil
}

func (service *EMPService) Sendbox(r *http.Request, args *NilParam, reply *[]objects.MetaMessage) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	*reply = localdb.GetBox(localdb.SENDBOX)
	return nil
}

func (service *EMPService) OpenMessage(r *http.Request, args *[]byte, reply *objects.FullMessage) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	var txidHash objects.Hash
	txidHash.FromBytes(*args)

	// Get Message from Database
	msg, err := localdb.GetMessageDetail(txidHash)
	if err != nil {
		return err
	}

	if msg.Encrypted == nil {
		*reply = *msg
		return nil
	}

	// If not decrypted, decrypt message and purge
	if msg.Decrypted == nil {
		recipient, err := localdb.GetAddressDetail(objects.MakeHash(encryption.StringToAddress(msg.MetaMessage.Recipient)))
		if err != nil {
			return err
		}

		if recipient.Privkey == nil {
			*reply = *msg
			return nil
		}

		// Decrypt Message
		decrypted := encryption.Decrypt(service.Config.Log, recipient.Privkey, msg.Encrypted)
		if len(decrypted) == 0 {
			*reply = *msg
			return nil
		}
		msg.Decrypted = new(objects.DecryptedMessage)
		msg.Decrypted.FromBytes(decrypted)

		// Update Sender
		detail := new(objects.AddressDetail)
		detail.Pubkey = msg.Decrypted.Pubkey[:]
		x, y := encryption.UnmarshalPubkey(detail.Pubkey)
		detail.Address = encryption.GetAddress(service.Config.Log, x, y)
		detail.String = encryption.AddressToString(detail.Address)

		localdb.AddUpdateAddress(detail)
		msg.MetaMessage.Sender = detail.String

		// Send Purge Request
		purge := new(objects.Purge)
		purge.Txid = msg.Decrypted.Txid

		service.Config.RecvQueue <- *objects.MakeFrame(objects.PURGE, objects.BROADCAST, purge)
		msg.MetaMessage.Purged = true

		localdb.AddUpdateMessage(msg, localdb.Contains(msg.MetaMessage.TxidHash))
	} else {
		if msg.MetaMessage.Purged == false && localdb.Contains(txidHash) == localdb.INBOX {
			msg.MetaMessage.Purged = true
			localdb.AddUpdateMessage(msg, localdb.Contains(msg.MetaMessage.TxidHash))
		}
	}

	*reply = *msg
	return nil
}
