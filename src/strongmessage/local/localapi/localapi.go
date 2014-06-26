package localapi

import (
	"fmt"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net"
	"time"
	"net/http"
	"strongmessage/api"
	"strongmessage/local/localdb"
	"strongmessage/objects"
	"strongmessage/db"
	"strongmessage/encryption"
	"strongmessage/network"
	"crypto/elliptic"
)

type StrongService struct {
	Config *api.ApiConfig
	Log    chan string
}

type NilParam struct{}

func (s *StrongService) Version(r *http.Request, args *NilParam, reply *objects.Version) error {
	*reply = *s.Config.LocalVersion
	return nil
}

func Initialize(log chan string, config *api.ApiConfig, port uint16) error {

	e := localdb.Initialize(log, config.LocalDB)

	if e != nil {
		return e
	}

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	service := new(StrongService)
	service.Config = config
	service.Log = log
	s.RegisterService(service, "StrongService")

	http.Handle("/", s)

	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log <- fmt.Sprintf("RPC Listen Error: %s", e)
		return e
	}

	go http.Serve(l, nil)

	go register(log, config)

	log <- fmt.Sprintf("Started RPC Server on: %s", fmt.Sprintf(":%d", port))
	return nil
}

// Handle Pubkey and Message Registration
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
			 */

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
					log <- fmt.Sprintf("Error updating local msg database... %s", err.Error())
					break
				}

				// Send Message and move to sendbox

				err = localdb.LocalDB.Exec("DELETE FROM outbox WHERE txid_hash=?", msg.TxidHash)
				if err != nil {
					log <- fmt.Sprintf("Error updating local outbox database... %s", err.Error())
					break
				}

				config.RecvChan <- *network.NewFrame("msg", msg.GetBytes(log))

				err = localdb.LocalDB.Exec("INSERT INTO sendbox VALUES(?, ?, ?, ?)", msg.TxidHash, msg.Timestamp, sender, recipient)
				if err != nil {
					log <- fmt.Sprintf("Error updating local msg database... %s", err.Error())
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
				log <- fmt.Sprintf("Error updating local msg database... %s", err.Error())
				break
			}
			err = localdb.LocalDB.Exec("INSERT INTO inbox VALUES(?, ?, NULL, ?)", message.TxidHash, message.Timestamp.Unix(), recipient)
			if err != nil {
				log <- fmt.Sprintf("Error updating local inbox database... %s", err.Error())
				break
			}
			localdb.Add(string(message.TxidHash), localdb.INBOX)
		}
	}
}

func getPubkey(log chan string, config *api.ApiConfig, addrHash, address []byte) []byte {
	payload := make([]byte, 0, 90)

	for s, err := localdb.LocalDB.Query("SELECT pubkey FROM addressbook WHERE hash=?", addrHash); err == nil; err = s.Next() {
		s.Scan(&payload)
		if len(payload) > 0 {
			return payload
		}
	}

	if db.Contains(string(addrHash)) == db.PUBKEY {
		payload, _ = db.GetPubkey(log, addrHash)
	}

	if len(payload) > 0 {
		var IV [16]byte
		for i := 0; i < 16; i++ {
			IV[i] = payload[i]
		}
		pubkey := encryption.SymmetricDecrypt(IV, address, payload[16:])
		pubkey = pubkey[:65]

		// Check public Key
		x, y := elliptic.Unmarshal(elliptic.P256(), pubkey)
		if x == nil {
			log <- "Decrypted Public Key Invalid"
			return nil
		}

		address2, _ := encryption.GetAddress(log, x, y)
		if string(address) != string(address2) {
			log <- "Decrypted Public Key doesn't match provided address!"
			return nil
		}

		// Add public key to local db
		err := localdb.LocalDB.Exec("UPDATE addressbook SET pubkey=? WHERE hash=?", pubkey, addrHash)
		if err != nil {
			log <- fmt.Sprintf("Error updating pubkey in localdb... %s", err)
		}
		return pubkey
	}

	config.RecvChan <- *network.NewFrame("pubkeyrq", addrHash)
	return nil
}
