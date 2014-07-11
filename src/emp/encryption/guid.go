package encryption

import (
	"crypto/rand"
	"fmt"
)

func GUID(bytes int) {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	fmt.Println(uuid)
}
