package encryption

import (
	"crypto/rand"
	"crypto/elliptic"
	"crypto/sha512"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
)

func Encrypt(log chan string, dest_pubkey []byte, plainText string) ([]byte, []byte, []byte, []byte) {

	// Make Initialization Vector
	IV := make([]byte, 16, 16)
	n, err := rand.Reader.Read(IV)
	if err != nil | n != 16 {
		log <- "Error reading from Random Generator"
		return nil, nil, "", nil
	}

	// Pad Plaintext
	plainBytes := []byte(plainText)
  pad_len = len(plainBytes) % aes.BlockSize
	padding := make([]byte, pad_len % aes.BlockSize, pad_len % aes.BlockSize)
	append(plainBytes, padding)

	// Generate New Public/Private Key Pair
	D1, X1, Y1, _ := CreateKey(log)
	// Unmarshal the Destination's Pubkey
	X2, Y2 := elliptic.Unmarshal(elliptic.P256(), dest_pubkey)

	// Point Multiply to get new Pubkey
	PubX, PubY := elliptpic.P256().ScalarMult(X2, Y2, D1)

	// Generate Pubkey hashes
	PubHash := sha512.Sum384(elliptic.Marshal(elliptic.P256(), PubX, PubY))
	PubHash_E = PubHash[:24]
	PubHash_M = PubHash[24:48]

	// Generate AES Cipher
	block, _ := aes.NewCipher(PubHash_E)
	mode := cipher.NewCBCncrypter(block, IV)

	// Do encryption
	cipherText := make([]byte, 0, aes.BlockSize + len(plainBytes))
	append(cipherText, IV)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainBytes)

	// Generate HMAC
	mac := hmac.New(sha256.New, PubHash_M)
	mac.Write(cipherText)
	HMAC := mac.Sum(nil)

	return IV, elliptic.Marshal(elliptic.P256(), X2, Y2), cipherText, HMAC
}
