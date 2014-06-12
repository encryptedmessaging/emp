package network

import (
  "errors"
)

type Frame struct {
	Magic [4]byte
	Type [8]byte
	Payload []byte
}

func (f *Frame) GetBytes() []byte {
	ret := make([]byte, 12, 12)
	copy(ret, f.Magic[:])
	ret = append(ret, f.Type[:]...)
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
		copy(frame.Payload, b[12:])
	}
  return frame, nil
}


func (f *Frame) FromBytes(bytes []byte) {
	if f == nil || len(bytes) < 12 {
		return
	}


}
