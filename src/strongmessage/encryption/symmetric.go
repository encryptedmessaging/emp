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
