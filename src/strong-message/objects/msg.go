package objects

import (
	"encoding/big"
	

type Msg struct {
	AddrHash []byte
	TxidHash []byte
	Timestamp time.Time
	IV []byte
	X big.Int
	Y big.Int
	Encrypted []byte
	MAC []byte
}

func (m *Msg) FromBytes(log chan string, data []byte) {
  var buffer bytes.Buffer
  enc := gob.NewDecoder(&buffer)
  err := enc.Decode(m)
  if err != nil {
    log <- "Decoding error."
    log <- err.Error()
  }
}

func (m *Msg) GetBytes(log chan string) []byte {
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

type MsgUnencrypted struct {
	Txid []byte
	
