package encoders

import (
	"fmt"
	//"math"
)

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
	Range         float64
	Name          string
	Radius        float64
	padding       int
	halfWidth     int
	rangeInternal float64
	clipInput     bool
	resolution    float64
	n             int
	//nInternal represents the output area excluding the possible padding on each
	nInternal int

	Verbosity int
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

/* Return the bit offset of the first bit to be set in the encoder output.
For periodic encoders, this can be a negative number when the encoded output
wraps around. */
func (se *ScalerEncoder) getFirstOnBit(input float64) int {

	//if input == SENTINEL_VALUE_FOR_MISSING_DATA:
	//	return [None]
	//else:

	if input < se.MinVal {
		//Don't clip periodic inputs. Out-of-range input is always an error
		if se.clipInput && !se.Periodic {

			if se.Verbosity > 0 {
				fmt.Printf("Clipped input %v=%d to minval %d", se.Name, input, se.MinVal)
			}
			input = se.MinVal
		} else {
			panic(fmt.Sprintf("Input %v less than range %v - %v", input, se.MinVal, se.MaxVal))
		}

		if se.Periodic {

			// Don't clip periodic inputs. Out-of-range input is always an error
			if input >= se.MaxVal {
				panic(fmt.Sprintf("input %v greater than periodic range %v - %v", input, se.MinVal, se.MaxVal))
			}

		} else {

			if input > se.MaxVal {
				if se.clipInput {
					if se.Verbosity > 0 {
						fmt.Printf("Clipped input %v=%v to maxval %v", se.Name, input, se.MaxVal)
					}
					input = se.MaxVal
				} else {
					panic(fmt.Sprintf("input %v greater than range (%v - %v)", input, se.MinVal, se.MaxVal))
				}
			}
		}
	}

	centerbin := 0

	if se.Periodic {
		centerbin = int((input-se.MinVal)*float64(se.nInternal)/se.Range) + se.padding
	} else {
		centerbin = int(((input-se.MinVal)+se.resolution/2)/se.resolution) + se.padding
	}

	// We use the first bit to be set in the encoded output as the bucket index
	minbin := centerbin - se.halfWidth
	return minbin
}

// func (se *ScalerEncoder) Encode(input float64, learn bool) output []bool {

// 	// if input is not None and not isinstance(input, numbers.Number):
// 	// raise TypeError("Expected a scalar input but got input of type %s" % type(input))

// 	// if type(input) is float and math.isnan(input):
// 	// input = SENTINEL_VALUE_FOR_MISSING_DATA
// 	// // Get the bucket index to use
// 	// bucketIdx = self._getFirstOnBit(input)[0]
// 	return []bool
// }

// func (se *ScalerEncoder) GetWidth() int {
// 	//return
// }
