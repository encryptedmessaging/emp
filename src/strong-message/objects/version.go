package objects

import (
	"time"
	"bytes"
  "encoding/gob"
)

type Version struct {
	Version uint32
  Timestamp time.Time
  UserAgent string
}

func (v *Version) FromBytes(log chan string, data []byte) {
  var buffer bytes.Buffer
	enc := gob.NewDecoder(&buffer)
	err := enc.Decode(v)
	if err != nil {
    log <- "Decoding error."
		log <- err.Error()
	}
}

func (v *Version) GetBytes(log chan string) []byte {
  var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(v)
	if err != nil {
    log <- "Encoding error!"
		log <- err.Error()
    return nil
	} else {
		return buffer.Bytes()
	}

}
