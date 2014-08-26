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
	String       string `json:"address"`
	Address      []byte `json:"address_bytes"`
	IsRegistered bool   `json:"registered"`
	IsSubscribed bool   `json:"subscribed"`
	Pubkey       []byte `json:"public_key"`
	Privkey      []byte `json:"private_key"`
	EncPrivkey   []byte `json:"encrypted_privkey"`
	Label        string `json:"address_label"`
}
