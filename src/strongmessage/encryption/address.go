package encryption

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"github.com/ThePiachu/Split-Vanity-Miner-Golang/src/pkg/ripemd160"
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
	}
	return priv, x, y
}

func GetAddress(log chan string, x, y *big.Int) ([]byte, string) {
	pubKey := elliptic.Marshal(elliptic.P256(), x, y)
	ripemd := ripemd160.New()

	sum := sha512.Sum384(pubKey)
	appender := make([]byte, sha512.Size384, sha512.Size384)
	for i := 0; i < sha512.Size384; i++ {
		appender[i] = sum[i]
	}	

	appender = ripemd.Sum(appender)
	address := make([]byte, 1, 1)

	// Version 0x01
	address[0] = 0x01
	address = append(address, appender...)

	sum = sha512.Sum384(address)
	sum = sha512.Sum384(address)


	for i := 0; i < 4; i++ {
		address = append(address, sum[i])
	}

	return address, base64.StdEncoding.EncodeToString(address)
}
