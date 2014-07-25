package encoders

import (
//"github.com/zacg/ints"
)

/*
 A value encoder takes a value and encodes it with a partial sparse representation
of bits.
*/
type ValueEncoder interface {
	//Width in bits
	GetWidth() int
	IsDelta() bool
	EncodeIntoArray(input interface{}) []bool
	GetSubEncoders() []Encoder
	GetName() string
	GetDescription() string
}

//Encodes multivariable input
type Encoder struct {
	Encoders []ValueEncoder
}

func (e *Encoder) Width() int {
	result := 0
	for _, val := range e.Encoders {
		result += val.GetWidth()
	}
	return 0
}
