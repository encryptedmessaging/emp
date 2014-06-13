package encryption

import (
	"bytes"
	"crypto/elliptic"
	"fmt"
	"testing"
)

func TestCrypt(t *testing.T) {
	log := make(chan string, 5)

	// Generate personal key
	priv, x, y := CreateKey(log)

	pub := elliptic.Marshal(elliptic.P256(), x, y)

	message := "If you see this, the test has passed!"

	enc := Encrypt(log, pub, message)

	plainBytes := Decrypt(log, priv, enc)
	plainBytes = bytes.Split(plainBytes, []byte{0})[0]
	fmt.Println(string(plainBytes))
	if message != string(plainBytes) {
		t.Fail()
	}
}

func TestSampleAddr(t *testing.T) {
	log := make(chan string, 5)

	// Generate Key
	_, x, y := CreateKey(log)

	byteAddr, strAddr := GetAddress(log, x, y)

	//Check lengths
	if len(byteAddr) != 25 || len(strAddr) != 33 {
		fmt.Println("Bad lengths: ", len(byteAddr), len(strAddr))
		t.Fail()
	}

}
