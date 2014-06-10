package encryption

import (
	"crypto/rand"
	"crypto/elliptic"
	"crypto/sha512"
	"math/big"
	"encoding/base64"
	"github.com/ThePiachu/Split-Vanity-Miner-Golang/src/pkg/ripemd160"
)

func CreateKey(log chan string) ([]byte, *big.Int, *big.Int) {
	priv, x, y, err := elliptic.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log <- "Key Generation Error"
	}
	return priv, x, y
}

func GetAddress(log chan string, x, y *big.Int) ([]byte, string) {
	pubKey := elliptic.Marshal(elliptic.P256(), x, y)
	ripemd := ripemd160.New()

	appender := ripemd.Sum(sha512.Sum384(pubKey))
	address := make([]byte, 1, 1)
	
	// Version 0x01
	address[0] = 0x01
	append(address, appender)
	append(address, sha512.Sum384(sha512.Sum384(address))[:4])

	return address, base64.StdEncoding.EncodeToString(address)
}
