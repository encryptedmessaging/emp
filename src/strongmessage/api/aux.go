package api

import (
	"quibit"
	"strongmessage/objects"
)

// Handle a Version Request or Reply
func fVERSION(config *ApiConfig, frame quibit.Frame, version *objects.Version) {

}

// Handle Peer List Requests or Replies
func fPEER(config *ApiConfig, frame quibit.Frame, nodeList *objects.NodeList) {

}

// Handle Object Vector Requests or Replies
func fOBJ(config *ApiConfig, frame quibit.Frame, obj *objects.Obj) {

}

// Handle Object Detail Requests
func fGETOBJ(config *ApiConfig, frame quibit.Frame, hash *objects.Hash) {

}

// Handle Public Key Request Broadcasts
func fPUBKEY_REQUEST(config *ApiConfig, frame quibit.Frame, pubHash *objects.Hash) {

}

// Handle Public Key Broadcasts
func fPUBKEY(config *ApiConfig, frame quibit.Frame, pubkey *objects.EncryptedPubkey) {

}

// Handle Encrypted Message Broadcasts
func fMSG(config *ApiConfig, frame quibit.Frame, msg *objects.Message) {

}

// Handle Purge Broadcasts
func fPURGE(config *ApiConfig, frame quibit.Frame, purge *objects.Purge) {

}
