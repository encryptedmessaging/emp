package objects

type EncryptedData struct {
	IV         [16]byte
	PublicKey  [65]byte
	CipherText []byte
	HMAC       [32]byte
}

const (
	ivLen = 16
	pubkeyLen = 65
	hmacLen = 32
	minLen = ivLen + pubkeyLen + hmacLen
)

func EncryptedFromBytes(b []byte) *EncryptedData {
	if len(b) < minLen {
		return nil
	}

	ret := new(EncryptedData)

	copy(ret.IV[:], b[:ivLen])
	copy(ret.PublicKey[:], b[ivLen:ivLen + pubkeyLen])
	copy(ret.CipherText, b[ivLen + pubkeyLen : len(b) - hmacLen])
	copy(ret.HMAC[:], b[len(b)-hmacLen:])

	return ret
}

func (e *EncryptedData) GetBytes() []byte {
	ret := make([]byte, 0, 0)
	ret = append(ret, e.IV[:]...)
	ret = append(ret, e.PublicKey[:]...)
	ret = append(ret, e.CipherText...)
	ret = append(ret, e.HMAC[:]...)
	return ret
}