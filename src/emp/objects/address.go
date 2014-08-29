/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package objects

type AddressDetail struct {
	String       string `json:"address"`           // String representation of address (33-35 characters)
	Address      []byte `json:"address_bytes"`     // Byte representation of address (25 bytes)
	IsRegistered bool   `json:"registered"`        // Whether EMPLocal is saving messages to this address.
	IsSubscribed bool   `json:"subscribed"`        // Whether EMPLocal is saving Publications from this address.
	Pubkey       []byte `json:"public_key"`        // Unencrypted 65-byte public key.
	Privkey      []byte `json:"private_key"`       // Unencrypted 32-byte private key.
	EncPrivkey   []byte `json:"encrypted_privkey"` // Encrypted private key (any length).
	Label        string `json:"address_label"`     // Human-readable label for this address.
}
