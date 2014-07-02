package encryption

import (
	"code.google.com/p/go.crypto/ripemd160"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"strconv"
	"math/big"
)

// This type is a placeholder for returns.  It hasn't been implemented yet.
type Address struct {
	PrivateKey []byte
	X          *big.Int
	Y          *big.Int
}

func CreateKey(log chan string) ([]byte, *big.Int, *big.Int) {
	priv, x, y, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log <- "Key Generation Error"
		return nil, nil, nil
	}
	return priv, x, y
}

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

func ValidateAddress(addr []byte) bool {
	if len(addr) != 25 {
		return false
	}
	ripe := addr[:21]
	sum := sha512.Sum384(ripe)
	sum = sha512.Sum384(sum[:])

	for i:=0; i < 4; i++ {
		if sum[i] != addr[i+21] {
			return false
		}
	}

	return true
}

func AddressToString(addr []byte) string {
	if !ValidateAddress(addr) {
		return ""
	}

	return strconv.Itoa(int(addr[0])) + base64.StdEncoding.EncodeToString(addr[1:])
}

func StringToAddress(addr string) []byte {
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
