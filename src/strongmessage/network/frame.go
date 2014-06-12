package objects

type Frame struct {
	Magic [4]byte
	Type [8]byte
	Payload []byte
}

func (f *Frame) GetBytes() []byte {
	ret := make([]byte, 12, 12)
	copy(ret, f.Magic[:])
	ret = append(ret, f.Type[:])
	ret = append(ret, f.Payload)

	return ret
}

func (f *Frame) FromBytes(bytes []byte) {
	if f == nil || len(bytes) < 12 {
		return
	}

	copy(f.Magic[:], bytes[:4])
	copy(f.Type, bytes[4:12])

	if len(bytes) > 12 {
		copy(f.Payload, bytes[12:])
	} else {
		f.Payload = nil
	}

}
