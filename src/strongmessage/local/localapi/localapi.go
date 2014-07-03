package localapi

import (
	"fmt"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net"
	"net/http"
	"strongmessage/api"
	"strongmessage/db"
	"strongmessage/encryption"
	"strongmessage/local/localdb"
	"strongmessage/objects"
)

type StrongService struct {
	Config *api.ApiConfig
}

type NilParam struct{}

func (s *StrongService) Version(r *http.Request, args *NilParam, reply *objects.Version) error {
	*reply = s.Config.LocalVersion
	return nil
}

func Initialize(config *api.ApiConfig) error {

	e := localdb.Initialize(config.Log, config.LocalDB)

	if e != nil {
		return e
	}

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	service := new(StrongService)
	service.Config = config
	s.RegisterService(service, "StrongService")

	http.Handle("/", s)

	l, e := net.Listen("tcp", fmt.Sprintf(":%d", config.RPCPort))
	if e != nil {
		config.Log <- fmt.Sprintf("RPC Listen Error: %s", e)
		return e
	}

	go http.Serve(l, nil)

	//go register(config)

	config.Log <- fmt.Sprintf("Started RPC Server on: %s", fmt.Sprintf(":%d", config.RPCPort))
	return nil
}

func Cleanup() {
	localdb.Cleanup()
}

// Handle Pubkey and Message Registration
/*
func register(log chan string, config *api.ApiConfig) {
	var addrHash []byte
	var message objects.Message

	for {
		select {
		case addrHash = <-config.PubkeyRegister:
			var address []byte
			/* Received Public Key!
			 * 1. Check if associated address exists
			 * 2. Decrypt and store public key
			 * 3. Check outbox for outgoing messages with recipient.
			 * 4. Foreach from 3: send message and move to sendbox.


			 // Check if key is registered
			 if localdb.Contains(string(addrHash)) != localdb.ADDRESS {
			 	break
			 }
			 s, err := localdb.LocalDB.Query("SELECT address FROM addressbook WHERE hash=?", addrHash)
			 if err != nil {
			 	break
			 }
			 s.Scan(&address)
			 if len(address) == 0 {
			 	break
			 }


			 // Decrypt and store public key
			 pubkey := getPubkey(log, config, addrHash, address)
			 if pubkey == nil {
			 	break
			 }
			 // Check outbox for outgoing messages with recipient.
			 for s, err := localdb.LocalDB.Query("SELECT msg.txid_hash, outbox.timestamp, msg.decrypted, outbox.sender, outbox.recipient FROM outbox INNER JOIN msg ON outbox.txid_hash=msg.txid_hash WHERE outbox.recipient=?", address); err == nil; err = s.Next() {
			 	msg := new(objects.Message)
			 	var timestamp int64
			 	var decrypted []byte
			 	var sender, recipient []byte
				s.Scan(&msg.TxidHash, &timestamp, &decrypted, &sender, &recipient)
				msg.AddrHash = addrHash
				msg.Timestamp = time.Unix(timestamp, 0)
				msg.Content = *encryption.Encrypt(log, pubkey, string(decrypted))
				err = localdb.LocalDB.Exec("UPDATE msg SET encrypted=? WHERE txid_hash=?", msg.Content.GetBytes(), msg.TxidHash)
				if err != nil {
					config.Log <- fmt.Sprintf("Error updating local msg database... %s", err.Error())
					break
				}

				// Send Message and move to sendbox

				err = localdb.LocalDB.Exec("DELETE FROM outbox WHERE txid_hash=?", msg.TxidHash)
				if err != nil {
					config.Log <- fmt.Sprintf("Error updating local outbox database... %s", err.Error())
					break
				}

				config.RecvChan <- *network.NewFrame("msg", msg.GetBytes(log))

				err = localdb.LocalDB.Exec("INSERT INTO sendbox VALUES(?, ?, ?, ?)", msg.TxidHash, msg.Timestamp, sender, recipient)
				if err != nil {
					config.Log <- fmt.Sprintf("Error updating local msg database... %s", err.Error())
					break
				}


			}

		case message = <-config.MessageRegister:
			if localdb.Contains(string(message.AddrHash)) != localdb.ADDRESS {
				break
			}

			// Check registration, then store message in inbox
			s, err := localdb.LocalDB.Query("SELECT registered, address from addressbook WHERE hash=?", message.AddrHash)
			if err != nil {
				break
			}
			var recipient []byte
			var isRegistered bool
			s.Scan(&isRegistered, &recipient)
			if !isRegistered {
				break
			}
			err = localdb.LocalDB.Exec("INSERT INTO msg VALUES(?, ?, NULL, 0)", message.TxidHash, message.Content.GetBytes())
			if err != nil {
				config.Log <- fmt.Sprintf("Error updating local msg database... %s", err.Error())
				break
			}
			err = localdb.LocalDB.Exec("INSERT INTO inbox VALUES(?, ?, NULL, ?)", message.TxidHash, message.Timestamp.Unix(), recipient)
			if err != nil {
				config.Log <- fmt.Sprintf("Error updating local inbox database... %s", err.Error())
				break
			}
			localdb.Add(string(message.TxidHash), localdb.INBOX)
		}
	}
}
*/

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
