package network

import (
	"errors"
)

type Frame struct {
	Peer    Peer  // Used for REP/REQ Pattern only
	Magic   [4]byte
	Type    [8]byte
	Payload []byte
}

func (f *Frame) GetBytes() []byte {
	ret := make([]byte, 12, 12)
	copy(ret, f.Magic[:])
	copy(ret[4:], f.Type[:])
	ret = append(ret, f.Payload...)

	return ret
}

func FrameFromBytes(b []byte) (Frame, error) {
	var frame Frame
	if len(b) < 12 {
		return frame, errors.New("Frame too short")
	}
	copy(frame.Magic[:], b[:4])
	copy(frame.Type[:], b[4:12])

	if len(b) > 12 {
		frame.Payload = make([]byte, len(b[12:]), cap(b[12:]))
		copy(frame.Payload, b[12:])
	}
	return frame, nil
}
