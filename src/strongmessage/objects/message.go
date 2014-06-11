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

func (m *Message) FromBytes(log chan string, data []byte) error {
	var buffer bytes.Buffer
	enc := gob.NewDecoder(&buffer)
	err := enc.Decode(m)
	if err != nil {
		log <- "Decoding error."
		log <- err.Error()
		return err
	}
	//placeholder until done
	er := errors.New("Not implemented")
	return er
}

func (m *Message) GetBytes(log chan string) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(m)
	if err != nil {
		log <- "Encoding error!"
		log <- err.Error()
		return nil, err
	} else {
		return buffer.Bytes(), nil
	}

}
