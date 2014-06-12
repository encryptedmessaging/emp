package objects


type EncryptedMessage struct {
  IV [16]byte
  PublicKey [65]byte
  CipherText []byte
  HMAC []byte
}
