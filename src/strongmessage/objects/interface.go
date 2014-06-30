package objects

import (
	"quibit"
)

type Serializer interface {
	GetBytes() []byte
	FromBytes([]byte) error
}

func MakeFrame(command, t uint8, payload Serializer) *quibit.Frame {
	frame := new(quibit.Frame)
	frame.Configure(payload.GetBytes(), command)
	frame.Header.Type = t

	return frame
}
