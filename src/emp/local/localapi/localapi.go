package localapi

import (
	"emp/api"
	"emp/db"
	"emp/encryption"
	"emp/local/localdb"
	"emp/objects"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net"
	"net/http"
	"time"
)

type EMPService struct {
	Config *api.ApiConfig
}

type NilParam struct{}

func (s *EMPService) Version(r *http.Request, args *NilParam, reply *objects.Version) error {
	if !basicAuth(s.Config, r) {
		s.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	*reply = s.Config.LocalVersion
	return nil
}

func basicAuth(config *api.ApiConfig, r *http.Request) bool {
	if config == nil || r == nil {
		return false
	}

	auth := r.Header.Get("Authorization")

	auth2 := "Basic " + base64.StdEncoding.EncodeToString([]byte(config.RPCUser+":"+config.RPCPass))

	return (auth == auth2)
}

func Initialize(config *api.ApiConfig) error {

	e := localdb.Initialize(config.Log, config.LocalDB)

	if e != nil {
		return e
	}

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	service := new(EMPService)
	service.Config = config
	s.RegisterService(service, "EMPService")

	// Register RPC Services
	http.Handle("/rpc", s)

	// Register JS Client
	http.Handle("/", http.FileServer(http.Dir(config.HttpRoot)))

	l, e := net.Listen("tcp", fmt.Sprintf(":%d", config.RPCPort))
	if e != nil {
		config.Log <- fmt.Sprintf("RPC Listen Error: %s", e)
		return e
	}

	go http.Serve(l, nil)

	go register(config)

	config.Log <- fmt.Sprintf("Started RPC Server on: %s", fmt.Sprintf(":%d", config.RPCPort))
	return nil
}

func Cleanup() {
	localdb.Cleanup()
}

// Handle Pubkey, Message, and Purge Registration
func register(config *api.ApiConfig) {
	var message objects.Message
	var txid [16]byte

	for {
		select {
		case pubHash := <-config.PubkeyRegister:

			// Check if pubkey is in database...
			pubkey := checkPubkey(config, pubHash)

			if pubkey == nil {
				break
			}

			outbox := localdb.GetBox(localdb.OUTBOX)
			for _, metamsg := range outbox {
				recvHash := objects.MakeHash([]byte(metamsg.Recipient))
				if string(pubHash.GetBytes()) == string(recvHash.GetBytes()) {
					// Send message and move to sendbox
					msg, err := localdb.GetMessageDetail(metamsg.TxidHash)
					if err != nil {
						config.Log <- err.Error()
						break
					}
					msg.Encrypted = encryption.Encrypt(config.Log, pubkey, string(msg.Decrypted.GetBytes()))
					msg.MetaMessage.Timestamp = time.Now().Round(time.Second)
					err = localdb.AddUpdateMessage(msg, localdb.SENDBOX)
					if err != nil {
						config.Log <- err.Error()
						break
					}

					sendMsg := new(objects.Message)
					sendMsg.Timestamp = msg.MetaMessage.Timestamp
					sendMsg.TxidHash = msg.MetaMessage.TxidHash
					sendMsg.AddrHash = recvHash
					sendMsg.Content = *msg.Encrypted

					config.RecvQueue <- *objects.MakeFrame(objects.MSG, objects.BROADCAST, sendMsg)
				}
			}

		case message = <-config.MessageRegister:
			// If address is registered, store message in inbox
			detail, err := localdb.GetAddressDetail(message.AddrHash)
			if err != nil {
				config.Log <- "Message address not in database..."
				break
			}
			if !detail.IsRegistered {
				config.Log <- "Message not for registered address..."
				break
			}

			config.Log <- "Registering new encrypted message..."

			msg := new(objects.FullMessage)
			msg.MetaMessage.TxidHash = message.TxidHash
			msg.MetaMessage.Timestamp = message.Timestamp
			msg.MetaMessage.Recipient = detail.String
			msg.Encrypted = &message.Content

			err = localdb.AddUpdateMessage(msg, localdb.INBOX)
			if err != nil {
				config.Log <- err.Error()
			}
		case message = <-config.PubRegister:
			// If address is registered, store message in inbox
			detail, err := localdb.GetAddressDetail(message.AddrHash)
			if err != nil {
				config.Log <- "Message address not in database..."
				break
			}
			if !detail.IsSubscribed {
				config.Log <- "Not Subscribed to Address..."
				break
			}

			config.Log <- "Registering new publication..."

			msg := new(objects.FullMessage)
			msg.MetaMessage.TxidHash = message.TxidHash
			msg.MetaMessage.Timestamp = message.Timestamp
			msg.MetaMessage.Sender = detail.String
			msg.MetaMessage.Recipient = "<Subscription Message>"
			msg.Encrypted = &message.Content

			msg.Decrypted = new(objects.DecryptedMessage)
			msg.Decrypted.FromBytes(encryption.DecryptPub(config.Log, detail.Pubkey, msg.Encrypted))

			err = localdb.AddUpdateMessage(msg, localdb.INBOX)
			if err != nil {
				config.Log <- err.Error()
			}
		case txid = <-config.PurgeRegister:
			// If Message in database, mark as purged
			detail, err := localdb.GetMessageDetail(objects.MakeHash(txid[:]))
			if err != nil {
				break
			}
			detail.MetaMessage.Purged = true
			err = localdb.AddUpdateMessage(detail, -1)
			if err != nil {
				config.Log <- fmt.Sprintf("Error registering purge: %s", err)
			}
		} // End select
	} // End for
} // End register

func checkPubkey(config *api.ApiConfig, addrHash objects.Hash) []byte {

	// First check local DB
	detail, err := localdb.GetAddressDetail(addrHash)
	if err != nil {
		// If not in database, won't be able to decrypt anyway!
		return nil
	}
	if len(detail.Pubkey) > 0 {
		if db.Contains(addrHash) != db.PUBKEY {
			enc := new(objects.EncryptedPubkey)

			enc.IV, enc.Payload, _ = encryption.SymmetricEncrypt(detail.Address, string(detail.Pubkey))
			enc.AddrHash = objects.MakeHash(detail.Address)

			config.RecvQueue <- *objects.MakeFrame(objects.PUBKEY, objects.BROADCAST, enc)
		}
		return detail.Pubkey
	}

	// If not there, check local database
	if db.Contains(addrHash) == db.PUBKEY {
		enc := db.GetPubkey(config.Log, addrHash)

		pubkey := encryption.SymmetricDecrypt(enc.IV, detail.Address, enc.Payload)
		pubkey = pubkey[:65]

		// Check public Key
		x, y := encryption.UnmarshalPubkey(pubkey)
		if x == nil {
			config.Log <- "Decrypted Public Key Invalid"
			return nil
		}

		address2 := encryption.GetAddress(config.Log, x, y)
		if string(detail.Address) != string(address2) {
			config.Log <- "Decrypted Public Key doesn't match provided address!"
			return nil
		}

		detail.Pubkey = pubkey
		err := localdb.AddUpdateAddress(detail)
		if err != nil {
			config.Log <- "Error adding pubkey to local database!"
			return nil
		}

		return pubkey
	}

	// If not there, send a pubkey request
	config.RecvQueue <- *objects.MakeFrame(objects.PUBKEY_REQUEST, objects.BROADCAST, &addrHash)
	return nil
}
