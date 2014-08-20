package localapi

import (
	"emp/encryption"
	"emp/local/localdb"
	"emp/objects"
	"errors"
	"fmt"
	"net/http"
	"quibit"
)

var logChan chan string

func (service *EMPService) ConnectionStatus(r *http.Request, args *NilParam, reply *int) error {
	*reply = quibit.Status()
	return nil
}

func (service *EMPService) GetLabel(r *http.Request, args *string, reply *string) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	var err error

	address := encryption.StringToAddress(*args)
	if len(address) != 25 {
		return errors.New(fmt.Sprintf("Invalid Address: %s", address))
	}

	addrHash := objects.MakeHash(address)

	detail, err := localdb.GetAddressDetail(addrHash)
	if err != nil {
		return err
	}

	if len(detail.Label) > 0 {
		*reply = detail.Label
	} else {
		*reply = *args
	}
	return nil
}

func (service *EMPService) CreateAddress(r *http.Request, args *NilParam, reply *objects.AddressDetail) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	// Create Address

	priv, x, y := encryption.CreateKey(service.Config.Log)
	reply.Privkey = priv
	if x == nil {
		return errors.New("Key Pair Generation Error")
	}

	reply.Pubkey = encryption.MarshalPubkey(x, y)

	reply.IsRegistered = true

	reply.Address = encryption.GetAddress(service.Config.Log, x, y)

	if reply.Address == nil {
		return errors.New("Could not create address, function returned nil.")
	}

	reply.String = encryption.AddressToString(reply.Address)

	// Add Address to Database
	err := localdb.AddUpdateAddress(reply)
	if err != nil {
		service.Config.Log <- fmt.Sprintf("Error Adding Address: ", err)
		return err
	}

	// Send Pubkey to Network
	encPub := new(objects.EncryptedPubkey)

	encPub.AddrHash = objects.MakeHash(reply.Address)

	encPub.IV, encPub.Payload, err = encryption.SymmetricEncrypt(reply.Address, string(reply.Pubkey))
	if err != nil {
		service.Config.Log <- fmt.Sprintf("Error Encrypting Pubkey: ", err)
		return nil
	}

	// Record Pubkey for Network
	service.Config.RecvQueue <- *objects.MakeFrame(objects.PUBKEY, objects.BROADCAST, encPub)
	return nil
}

func (service *EMPService) GetAddress(r *http.Request, args *string, reply *objects.AddressDetail) error {

	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	var err error

	address := encryption.StringToAddress(*args)
	if len(address) != 25 {
		return errors.New(fmt.Sprintf("Invalid Address: %s", address))
	}

	addrHash := objects.MakeHash(address)

	detail, err := localdb.GetAddressDetail(addrHash)
	if err != nil {
		return err
	}

	// Check for pubkey
	if len(detail.Pubkey) == 0 {
		detail.Pubkey = checkPubkey(service.Config, objects.MakeHash(detail.Address))
	}

	*reply = *detail

	return nil
}

func (service *EMPService) AddUpdateAddress(r *http.Request, args *objects.AddressDetail, reply *NilParam) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	err := localdb.AddUpdateAddress(args)
	if err != nil {
		return err
	}

	checkPubkey(service.Config, objects.MakeHash(args.Address))

	return nil
}

func (service *EMPService) ListAddresses(r *http.Request, args *bool, reply *([][2]string)) error {
	if !basicAuth(service.Config, r) {
		service.Config.Log <- fmt.Sprintf("Unauthorized RPC Request from: %s", r.RemoteAddr)
		return errors.New("Unauthorized")
	}

	strs := localdb.ListAddresses(*args)
	*reply = strs
	return nil
}
