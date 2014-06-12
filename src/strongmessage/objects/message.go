package objects

import (
  "bytes"
  "encoding/gob"
)

type Message struct {
  Magic uint32
  Type string
  Payload []byte
}


func (m *Message) FromBytes(log chan string, data []byte) {
  var buffer bytes.Buffer
  enc := gob.NewDecoder(&buffer)
  err := enc.Decode(m)
  if err != nil {
    log <- "Decoding error."
    log <- err.Error()
  }
}

func (m *Message) GetBytes(log chan string) []byte {
  var buffer bytes.Buffer
  enc := gob.NewEncoder(&buffer)
  err := enc.Encode(m)
  if err != nil {
    log <- "Encoding error!"
    log <- err.Error()
    return nil
  } else {
    return buffer.Bytes()
  }

}
