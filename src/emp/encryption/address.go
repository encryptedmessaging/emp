/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package encryption

import (
	"golang.org/x/crypto/ripemd160"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"math/big"
	"strconv"
)

// Create a new Public-Private ECC-256 Keypair.
func CreateKey(log chan string) ([]byte, *big.Int, *big.Int) {
	priv, x, y, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log <- "Key Generation Error"
		return nil, nil, nil
	}
	return priv, x, y
}

// Convert public key to uncompressed 65-byte slice (2 32-bit integers with prefix 0x04)
func MarshalPubkey(x, y *big.Int) []byte {
	return elliptic.Marshal(elliptic.P256(), x, y)
}

// Convert 65-byte slice as created by MarshalPubkey() into an ECC-256 Public Key.
func UnmarshalPubkey(data []byte) (x, y *big.Int) {
	return elliptic.Unmarshal(elliptic.P256(), data)
}

func GetCurve() elliptic.Curve {
	return elliptic.P256()
}

// Convert ECC-256 Public Key to an EMP address (raw 25 bytes).
func GetAddress(log chan string, x, y *big.Int) []byte {
	pubKey := elliptic.Marshal(elliptic.P256(), x, y)
	ripemd := ripemd160.New()

	sum := sha512.Sum384(pubKey)
	sumslice := make([]byte, sha512.Size384, sha512.Size384)
	for i := 0; i < sha512.Size384; i++ {
		sumslice[i] = sum[i]
	}

	ripemd.Write(sumslice)
	appender := ripemd.Sum(nil)
	appender = appender[len(appender)-20:]
	address := make([]byte, 1, 1)

	// Version 0x01
	address[0] = 0x01
	address = append(address, appender...)

	sum = sha512.Sum384(address)
	sum = sha512.Sum384(sum[:])

	for i := 0; i < 4; i++ {
		address = append(address, sum[i])
	}

	return address
}

// Determine if address is valid (checksum is correct, correct length, and it starts with a 1-byte number).
func ValidateAddress(addr []byte) bool {
	if len(addr) != 25 {
		return false
	}
	ripe := addr[:21]
	sum := sha512.Sum384(ripe)
	sum = sha512.Sum384(sum[:])

	for i := 0; i < 4; i++ {
		if sum[i] != addr[i+21] {
			return false
		}
	}

	return true
}

// Converts 25-byte address to String representation.
func AddressToString(addr []byte) string {
	if !ValidateAddress(addr) {
		return ""
	}

	return strconv.Itoa(int(addr[0])) + base64.StdEncoding.EncodeToString(addr[1:])
}

// Converts String representation to 25-byte address.
func StringToAddress(addr string) []byte {
	if len(addr) < 2 {
		return nil
	}
	data, err := base64.StdEncoding.DecodeString(addr[1:])
	if err != nil {
		return nil
	}
	version := make([]byte, 1, 1)
	version[0] = byte(addr[0] - 48)
	address := append(version, data...)
	if !ValidateAddress(address) {
		return nil
	}
	return address
}
