package objects

import (
  "bytes"
  "encoding/gob"
  "time"
)

type Message struct {
  AddrHash  []byte
  TxidHash  []byte
  Timestamp time.Time
  Content   EncryptedData
}

// Lets allow for multiple datatypes even if we don't support them in the first
// itteration.
type MessageUnencrypted struct {
  Txid      []byte
  SendAddr  []byte
  Timestamp time.Time
  DataType  string
  Data      []byte
}

func MessageFromBytes(log chan string, data []byte) (Message, error) {
  buffer := bytes.NewBuffer(data)
  decoder := gob.NewDecoder(buffer)
  var message Message
  err := decoder.Decode(&message)
  if err != nil {
    log <- "Decoding error."
    log <- err.Error()
    return Message{}, err
  } else {
    return message, nil
  }

}

func (m *Message) GetBytes(log chan string) []byte {
  var buffer bytes.Buffer
  gob.Register(Message{})
  enc := gob.NewEncoder(&buffer)
  err := enc.Encode(&m)
  if err != nil {
    log <- "Encoding error!"
    log <- err.Error()
    return nil
  } else {
    return buffer.Bytes()
  }
}
