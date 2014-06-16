package network

import (
	"errors"
	"encoding/gob"
	"bytes"
)

type Frame struct {
	Peer    *Peer  // Used for REP/REQ Pattern only
	Magic   [4]byte
	Type    string
	Payload []byte
}

func (f *Frame) GetBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(f)
	if err != nil {
		return nil
	} else {
		return buffer.Bytes()
	}
}

func FrameFromBytes(b []byte) (Frame, error) {
	var frame Frame
	if len(b) < 12 {
		return frame, errors.New("Frame too short")
	}

	var buffer bytes.Buffer
	enc := gob.NewDecoder(&buffer)
	err := enc.Decode(&frame)
	if err != nil {
		return frame, err
	}

	return frame, nil
}
