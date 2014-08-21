/**
    Copyright 2014 JARST, LLC
    
    This file is part of EMP.

    EMP is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
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
