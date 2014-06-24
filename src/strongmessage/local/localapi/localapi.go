package localapi

import (
	"fmt"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net"
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

	log <- fmt.Sprintf("Started RPC Server on: %s", fmt.Sprintf(":%d", port))
	return nil
}

func getPubkey(log chan string, config *api.ApiConfig, addrHash []byte, address []byte) []byte {
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
