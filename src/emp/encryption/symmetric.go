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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func SymmetricEncrypt(key []byte, plainText string) ([aes.BlockSize]byte, []byte, error) {

	// Make Initialization Vector
	var IV [aes.BlockSize]byte
	n, err := rand.Reader.Read(IV[:])
	if err != nil || n != 16 {
		return IV, nil, err
	}

	// Pad Plaintext
	plainBytes := []byte(plainText)

	pad_len := aes.BlockSize - (len(plainBytes) % aes.BlockSize)

	padding := make([]byte, pad_len, pad_len)
	plainBytes = append(plainBytes, padding...)

	// Generate AES Cipher

	if len(key) == 25 {
		key = append(key, make([]byte, 7, 7)...)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Uh Oh: ", err)
		return IV, nil, err
	}
	mode := cipher.NewCBCEncrypter(block, IV[:])

	// Do encryption
	cipherText := make([]byte, len(plainBytes), len(plainBytes))
	mode.CryptBlocks(cipherText, plainBytes)

	return IV, cipherText, nil
}

func SymmetricDecrypt(IV [aes.BlockSize]byte, key, cipherText []byte) []byte {
	plainText := make([]byte, len(cipherText), len(cipherText))

	if len(key) == 25 {
		key = append(key, make([]byte, 7, 7)...)
	}

	// Generate AES Cipher
	block, _ := aes.NewCipher(key)
	mode := cipher.NewCBCDecrypter(block, IV[:])

	// Do decryption
	plainText = make([]byte, len(cipherText), len(cipherText))
	mode.CryptBlocks(plainText, cipherText[:])

	return plainText
}
