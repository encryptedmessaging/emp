package api

import (
	"quibit"
	"strongmessage/objects"
)

// Handle a Version Request or Reply
func fVERSION(config *ApiConfig, frame quibit.Frame, version *objects.Version) {
	// Verify Protcol Version, else Disconnect

	// Verify Timestamp (5 minute window), else Disconnect

	// If backbone node, verify IP, then add to master node list

	// If a REQUEST, send local version as a REPLY

	// If a REPLY, send a peer list as a REQUEST

}

// Handle Peer List Requests or Replies
func fPEER(config *ApiConfig, frame quibit.Frame, nodeList *objects.NodeList) {
	// Copy master Node List to slave Node LIst

	// For each Node in incoming Node List

		// If not in master node list, connect and add to master node list

		// Else, remove from slave Node List

	// If a REQUEST, send slave Node List as a REPLY

	// if a REPLY, send local Object List as REQUEST

}

// Handle Object Vector Requests or Replies
func fOBJ(config *ApiConfig, frame quibit.Frame, obj *objects.Obj) {
	// If a REQUEST, send local object list as REPLY

	// For each object in object list:
		// If object not stored locally, send GETOBJ REQUEST
}

// Handle Object Detail Requests
func fGETOBJ(config *ApiConfig, frame quibit.Frame, hash *objects.Hash) {
	// If object stored locally, send object as a REPLY

}

// Handle Public Key Request Broadcasts
func fPUBKEY_REQUEST(config *ApiConfig, frame quibit.Frame, pubHash *objects.Hash) {
	// Check Hash in Object List
	
	// If request is Not in List, store the request
		// If a BROADCAST, send out another BROADCAST

	// If request is a Public Key in List:
		// Send out the PUBKEY as a BROADCAST
}

// Handle Public Key Broadcasts
func fPUBKEY(config *ApiConfig, frame quibit.Frame, pubkey *objects.EncryptedPubkey) {
	// Check Hash in Object List

	// If request is a PUBKEY_REQUEST or NOT in List, store the request
		// If a BROADCAST, send out another BROADCAST
}

// Handle Encrypted Message Broadcasts
func fMSG(config *ApiConfig, frame quibit.Frame, msg *objects.Message) {
	// Check Hash in Object List

	// Verify timestamp is newer than 365 days

	// If Not in List, store the request
		// If a BROADCAST, send out another BROADCAST


}

// Handle Purge Broadcasts
func fPURGE(config *ApiConfig, frame quibit.Frame, purge *objects.Purge) {
	// Check Hash in Object List

	// If Not in List or MSG, delete the MSG and store the request
		// If a BROADCAST, send out another BROADCAST

}
