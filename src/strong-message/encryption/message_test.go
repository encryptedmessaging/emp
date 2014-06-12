package encryption

import (
	"testing"
	"fmt"
	"crypto/elliptic"
	"bytes"
)

func TestCrypt(t *testing.T) {
	log := make(chan string, 5)
	
	// Generate personal key
	priv, x, y := CreateKey(log)

	pub := elliptic.Marshal(elliptic.P256(), x, y)

	message := "If you see this, the test has passed!"

	iv, pub2, cipher, MAC := Encrypt(log, pub, message)
	
	plainBytes := Decrypt(log, priv, iv, pub2, cipher, MAC)
	plainBytes = bytes.Split(plainBytes, []byte{0})[0]
	fmt.Println(string(plainBytes))
	if message != string(plainBytes) {
		t.Fail()
	}
}
