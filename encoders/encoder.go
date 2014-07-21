package encoders

import ()

/*
 An encoder takes a value and encodes it with a partial sparse representation
of bits. The Encoder interface implements:
*/
type Encoder interface {
	//Width in bits
	GetWidth() int
	IsDelta() bool
	EncodeIntoArray(input interface{}) []bool
}
