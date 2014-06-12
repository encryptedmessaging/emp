package objects


type EncryptedData struct {
  IV         [16]byte
  PublicKey  [65]byte
  CipherText []byte
  HMAC       [32]byte
}
