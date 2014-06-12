package encryption

import (
	"crypto/aes"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/cipher"
)

func Encrypt(log chan string, dest_pubkey []byte, plainText string) objects.EncryptedData {

	// Make Initialization Vector
	IV := make([]byte, 16, 16)
	n, err := rand.Reader.Read(IV)
	if err != nil || n != 16 {
		log <- "Error reading from Random Generator"
		return nil, nil, nil, nil
	}

	// Pad Plaintext
	plainBytes := []byte(plainText)

  pad_len := aes.BlockSize - (len(plainBytes) % aes.BlockSize)

	padding := make([]byte, pad_len, pad_len)
	plainBytes = append(plainBytes, padding...)

	// Generate New Public/Private Key Pair
	D1, X1, Y1 := CreateKey(log)
	// Unmarshal the Destination's Pubkey
	X2, Y2 := elliptic.Unmarshal(elliptic.P256(), dest_pubkey)

	// Point Multiply to get new Pubkey
	PubX, PubY := elliptic.P256().ScalarMult(X2, Y2, D1)

	// Generate Pubkey hashes
	PubHash := sha512.Sum384(elliptic.Marshal(elliptic.P256(), PubX, PubY))
	PubHash_E := PubHash[:24]
	PubHash_M := PubHash[24:48]

	// Generate AES Cipher
	block, _ := aes.NewCipher(PubHash_E)
	mode := cipher.NewCBCEncrypter(block, IV)

	// Do encryption
	cipherText := make([]byte, len(plainBytes), len(plainBytes))
	mode.CryptBlocks(cipherText, plainBytes)

	// Generate HMAC
	mac := hmac.New(sha256.New, PubHash_M)
	mac.Write(cipherText)
	HMAC := mac.Sum(nil)
	encrypted_data = objects.EncryptedData{IV: IV, PublicKey: elliptic.Marshal(elliptic.P256(), X2, Y2), CipherText: cipherText, HMAC: HMAC}
	return encrypted_data

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
	PubX, PubY := elliptic.P256().ScalarMult(X2, Y2, privKey)

	// Generate Pubkey hashes
	PubHash := sha512.Sum384(elliptic.Marshal(elliptic.P256(), PubX, PubY))
	PubHash_E = PubHash[:24]
	PubHash_M = PubHash[24:48]
  
  // Generate Pubkey hashes 
  PubHash := sha512.Sum384(elliptic.Marshal(elliptic.P256(), PubX, PubY))
  PubHash_E := PubHash[:24]
  PubHash_M := PubHash[24:48]

	// Check HMAC
	if !checkMAC(cipherText, HMAC, PubHash_M) {
		log <- "Invalid HMAC Message"
		return nil
	}

	// Generate AES Cipher
	block, _ := aes.NewCipher(PubHash_E)
	mode := cipher.NewCBCDecrypter(block, IV)

  // Do decryption
  plainText := make([]byte, len(cipherText), len(cipherText))
  mode.CryptBlocks(plainText, cipherText)

	return plainText
}
