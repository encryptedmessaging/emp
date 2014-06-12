package objects

// The cipher text should start with content-type.

type EncryptedData struct {
	IV         [16]byte
	PublicKey  [65]byte
	CipherText []byte
	HMAC       []byte
}
