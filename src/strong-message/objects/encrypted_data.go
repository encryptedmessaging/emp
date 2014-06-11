package objects

// The cipher text should start with content-type.

type EncryptedData struct {
  IV []byte
  PublicKey []byte
  CipherText []byte
  HMAC []byte
}
