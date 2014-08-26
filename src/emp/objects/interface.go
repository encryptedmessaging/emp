/**
    Copyright 2014 JARST, LLC.
    
    This file is part of EMP.

    EMP is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the included
    LICENSE file for more details.
**/

package objects

import (
	"quibit"
)

type Serializer interface {
	GetBytes() []byte
	FromBytes([]byte) error
}

type NilPayload bool

func (n *NilPayload) GetBytes() []byte {
	return nil
}

func (n *NilPayload) FromBytes(b []byte) error {
	return nil
}

func MakeFrame(command, t uint8, payload Serializer) *quibit.Frame {
	frame := new(quibit.Frame)
	frame.Configure(payload.GetBytes(), command, t)
	frame.Peer = ""

	return frame
}
