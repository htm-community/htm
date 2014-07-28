package encoders

import ()

/*
 n -- The number of bits in the output. Must be greater than or equal to w

radius -- Two inputs separated by more than the radius have non-overlapping
representations. Two inputs separated by less than the radius will
in general overlap in at least some of their bits. You can think
of this as the radius of the input.

resolution -- Two inputs separated by greater than, or equal to the resolution are guaranteed
to have different representations.
*/
type ScalerOutputType int

const (
	N          ScalerOutputType = 1
	Radius     ScalerOutputType = 2
	Resolution ScalerOutputType = 3
)

/*
 A scalar encoder encodes a numeric (floating point) value into an array
of bits. The output is 0's except for a contiguous block of 1's. The
location of this contiguous block varies continuously with the input value.

The encoding is linear. If you want a nonlinear encoding, just transform
the scalar (e.g. by applying a logarithm function) before encoding.
It is not recommended to bin the data as a pre-processing step, e.g.
"1" = $0 - $.20, "2" = $.21-$0.80, "3" = $.81-$1.20, etc. as this
removes a lot of information and prevents nearby values from overlapping
in the output. Instead, use a continuous transformation that scales
the data (a piecewise transformation is fine).
*/
type ScalerEncoder struct {
	Width         int
	MinVal        float64
	MaxVal        float64
	Periodic      bool
	OutputType    ScalerOutputType
	padding       int
	halfWidth     int
	rangeInternal float64
}

func NewScalerEncoder(width int) *ScalerEncoder {
	se := new(ScalerEncoder)
	if width%2 == 0 {
		panic("Width must be an odd number.")
	}

	se.halfWidth = width / 2
	if !se.Periodic {
		se.padding = se.halfWidth
	}

	return se
}

func (se *ScalerEncoder) Encode(input []string, output []bool) {

}
