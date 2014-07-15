package encryption

import "errors"

type EncryptedMessage struct {
	IV         [16]byte
	PublicKey  [65]byte
	CipherText []byte
	HMAC       [32]byte
}

const (
	ivLen     = 16
	pubkeyLen = 65
	hmacLen   = 32
	minLen    = ivLen + pubkeyLen + hmacLen
)

func (ret *EncryptedMessage) FromBytes(b []byte) error {
	if len(b) < minLen {
		return errors.New("Bytes too short to create EncryptedMessage object.")
	}
	if ret == nil {
		return errors.New("Can't fill nil object.")
	}

	copy(ret.IV[:], b[:ivLen])
	copy(ret.PublicKey[:], b[ivLen:ivLen+pubkeyLen])
	ret.CipherText = append(ret.CipherText, b[ivLen+pubkeyLen:len(b)-hmacLen]...)
	copy(ret.HMAC[:], b[len(b)-hmacLen:])

	return nil
}

func (e *EncryptedMessage) GetBytes() []byte {
	if e == nil {
		return nil
	}
	ret := make([]byte, 0, 0)
	ret = append(ret, e.IV[:]...)
	ret = append(ret, e.PublicKey[:]...)
	ret = append(ret, e.CipherText...)
	ret = append(ret, e.HMAC[:]...)

	return ret
}
