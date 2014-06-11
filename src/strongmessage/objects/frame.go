package objects

import (
	"bytes"
	"encoding/gob"
	"errors"
)

type Frame struct {
	Magic   uint32
	Type    string
	Payload []byte
}

func (m *Frame) FromBytes(log chan string, data []byte) error {
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

func (m *Frame) GetBytes(log chan string) ([]byte, error) {
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
