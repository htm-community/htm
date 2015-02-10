package encoders

import (
	"fmt"
	//"math"
	"github.com/zacg/htm/utils"
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

/*
 Returns bucket index for given input
*/
func (se *ScalerEncoder) getBucketIndices(input float64) []int {

	minbin := se.getFirstOnBit(input)
	var bucketIdx int

	// For periodic encoders, the bucket index is the index of the center bit
	if se.Periodic {
		bucketIdx = minbin + se.halfWidth
		if bucketIdx < 0 {
			bucketIdx += se.n
		}
	} else {
		// for non-periodic encoders, the bucket index is the index of the left bit
		bucketIdx = minbin
	}

	return []int{bucketIdx}
}

func (se *ScalerEncoder) Encode(input float64, learn bool) (output []bool) {

	// Get the bucket index to use
	bucketIdx := se.getFirstOnBit(input)

	//if len(bucketIdx) {
	//This shouldn't get hit
	//	panic("Missing input value")
	//TODO output[0:self.n] = 0 TODO: should all 1s, or random SDR be returned instead?
	//} else {
	// The bucket index is the index of the first bit to set in the output
	output = make([]bool, se.n)
	minbin := bucketIdx
	maxbin := minbin + 2*se.halfWidth

	if se.Periodic {

		// Handle the edges by computing wrap-around
		if maxbin >= se.n {
			bottombins := maxbin - se.n + 1
			utils.FillSliceRangeBool(output, true, 0, bottombins)
			maxbin = se.n - 1
		}
		if minbin < 0 {
			topbins := -minbin
			//output = output[se.n - topbins:se.n]
			utils.FillSliceRangeBool(output, true, se.n-topbins, se.n)
			minbin = 0
		}

	}

	if minbin < 0 {
		panic("invalid minbin")
	}
	if maxbin >= se.n {
		panic("invalid maxbin")
	}

	// set the output (except for periodic wraparound)
	utils.FillSliceRangeBool(output, true, minbin, maxbin+1)

	if se.Verbosity >= 2 {
		fmt.Println("input:", input)
		fmt.Printf("range: %v - %v \n", se.MinVal, se.MaxVal)
		fmt.Printf("n: %v width: %v resolution: %v \n", se.n, se.Width, se.resolution)
		fmt.Printf("radius: %v periodic: %v \n", se.Radius, se.Periodic)
		fmt.Printf("output: %v \n", output)
	}

	//}

	return output
}
