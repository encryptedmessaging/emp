package objects


type EncryptedMessage struct {
  IV []byte
  PublicKey []byte
  CipherText []byte
  HMAC []byte
}
