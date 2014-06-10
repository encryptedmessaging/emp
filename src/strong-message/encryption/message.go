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
	padding := make([]byte, (len(plainBytes) % aes.BlockSize) % aes.BlockSize, (len(plainBytes) % aes.BlockSize) % aes.BlockSize)
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
	mode := cipher.NewCBCEncrypter(block, IV)

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

// checkMAC returns true if messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func Decrypt(log chan string, privKey, IV, pubKey, cipherText, HMAC []byte) []byte {
	// Unmarshal the Sender's Pubkey
	X2, Y2 := elliptic.Unmarshal(elliptic.P256(), pubKey)

	// Point Multiply to get the new Pubkey
	PubX, PubY := elliptpic.P256().ScalarMult(X2, Y2, D1)

  // Generate Pubkey hashes 
  PubHash := sha512.Sum384(elliptic.Marshal(elliptic.P256(), PubX, PubY))
  PubHash_E = PubHash[:24]
  PubHash_M = PubHash[24:48]

	// Check HMAC
	if !checkMAC(cipherText, HMAC, PubHash_M) {
		log <- "Invalid HMAC Message"
		return nil
	}

	// Generate AES Cipher
  block, _ := aes.NewCipher(PubHash_E)
  mode := cipher.NewCBCDecrypter(block, IV)

  // Do decryption
  plainText := make([]byte, 0, len(cipherText))
  mode.CryptBlocks(plainText, cipherText)

	return plainText
}
