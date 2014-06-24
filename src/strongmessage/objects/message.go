package objects

import (
	"bytes"
	"encoding/gob"
	"encoding/binary"
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
	Signature []byte
}

const (
	unMinLen = 16+25+8
)

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

func DecryptedFromBytes(b []byte) *MessageUnencrypted {
	if len(b) < unMinLen {
		return nil
	}

	ret := new(MessageUnencrypted)

	ret.Txid = append(ret.Txid, b[:16]...)
	ret.SendAddr = append(ret.SendAddr, b[16:16+25]...)
	ret.Timestamp = time.Unix(int64(binary.BigEndian.Uint64(b[16+25:unMinLen])), 0)

	for i := unMinLen; i < len(b); i++ {
		if b[i] == 0x00 {
			typeStr := make([]byte, i - (unMinLen), i - (unMinLen))
			copy(typeStr, b[unMinLen:i])
			ret.DataType = string(typeStr)
			ret.Data = append(ret.Data, b[i:len(b)-65]...)
			ret.Signature = append(ret.Signature, b[len(b)-65:]...)
			return ret
		}
	}

	return nil
}

func (e *MessageUnencrypted) GetBytes() []byte {
	ret := make([]byte, 0, 0)
	ret = append(ret, e.Txid...)
	ret = append(ret, e.SendAddr...)

	timestmp := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(timestmp, uint64(e.Timestamp.Unix()))
	ret = append(ret, timestmp...)

	ret = append(ret, e.DataType...)
	ret = append(ret, 0x00)

	ret = append(ret, e.Data...)
	ret = append(ret, e.Signature...)

	return ret
}