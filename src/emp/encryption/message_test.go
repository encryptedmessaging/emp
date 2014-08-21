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

package encryption

import (
	"bytes"
	"crypto/elliptic"
	"fmt"
	"testing"
)

func TestCryptPub(t *testing.T) {
	log := make(chan string, 5)

	// Generate personal key
	priv, x, y := CreateKey(log)

	pub := elliptic.Marshal(elliptic.P256(), x, y)

	message := "If you see this, the test has passed!"

	enc := EncryptPub(log, priv, message)

	plainBytes := DecryptPub(log, pub, enc)
	plainBytes = bytes.Split(plainBytes, []byte{0})[0]
	if message != string(plainBytes) {
		t.Fail()
	}
}

func TestCrypt(t *testing.T) {
	log := make(chan string, 5)

	// Generate personal key
	priv, x, y := CreateKey(log)

	pub := elliptic.Marshal(elliptic.P256(), x, y)

	message := "If you see this, the test has passed!"

	enc := Encrypt(log, pub, message)

	plainBytes := Decrypt(log, priv, enc)
	plainBytes = bytes.Split(plainBytes, []byte{0})[0]
	if message != string(plainBytes) {
		t.Fail()
	}
}

func TestSampleAddr(t *testing.T) {
	log := make(chan string, 5)

	// Generate Key
	_, x, y := CreateKey(log)

	byteAddr := GetAddress(log, x, y)

	//Check lengths
	if len(byteAddr) != 25 {
		fmt.Println("Bad length: ", len(byteAddr))
		t.Fail()
	}

	if !ValidateAddress(byteAddr) {
		fmt.Println("Address validation falied!")
		t.Fail()
	}

	if string(StringToAddress(AddressToString(byteAddr))) != string(byteAddr) {
		fmt.Println("Error in the address/string conversion functions: ", StringToAddress(AddressToString(byteAddr)))
		t.Fail()
	}

}
