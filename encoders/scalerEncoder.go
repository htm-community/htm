package encoders

import (
	"fmt"
	"github.com/zacg/htm/utils"
	"math"
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

type ScalerEncoderParams struct {
	Width      int
	MinVal     float64
	MaxVal     float64
	Periodic   bool
	OutputType ScalerOutputType
	Range      float64
	Resolution float64
	Name       string
	Radius     float64
	ClipInput  bool
	Verbosity  int
	N          int
}

func NewScalerEncoderParams(width int, minVal float64, maxVal float64) *ScalerEncoderParams {
	p := new(ScalerEncoderParams)

	p.Width = width
	p.MinVal = minVal
	p.MaxVal = maxVal
	p.N = 0
	p.Radius = 0
	p.Resolution = 0
	p.Name = ""
	p.Verbosity = 0
	p.ClipInput = false

	return p
}

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
	params        *ScalerEncoderParams
	padding       int
	halfWidth     int
	rangeInternal float64

	//nInternal represents the output area excluding the possible padding on each
	nInternal int
}

func NewScalerEncoder(params *ScalerEncoderParams) *ScalerEncoder {
	se := new(ScalerEncoder)
	if params.Width%2 == 0 {
		panic("Width must be an odd number.")
	}

	se.halfWidth = params.Width / 2
	if !params.Periodic {
		se.padding = se.halfWidth
	}

	return se
}

/*
	helper used to inititalize the encoder
*/
func (se *ScalerEncoder) init(width int, minval float64, maxval float64, n int,
	radius float64, resolution float64) {
	//handle 3 diff ways of representation

	if n != 0 {
		//crutches ;(
		if radius == 0 {
			panic("radius is 0")
		}
		if resolution == 0 {
			panic("radius is 0")
		}
		if n <= width {
			panic("n less than width")
		}

		se.params.N = n

		//if (minval is not None and maxval is not None){

		if !se.params.Periodic {
			se.params.Resolution = se.rangeInternal / float64(se.params.N-se.params.Width)
		} else {
			se.params.Resolution = se.rangeInternal / float64(se.params.N)
		}

		se.params.Radius = float64(se.params.Width) * se.params.Resolution

		if se.params.Periodic {
			se.params.Range = se.rangeInternal
		} else {
			se.params.Range = se.rangeInternal + se.params.Resolution
		}

	} else { //n == 0
		if radius != 0 {
			if resolution != 0 {
				panic("resolution not 0")
			}
			se.params.Radius = radius
			se.params.Resolution = se.params.Radius / float64(width)
		} else if resolution != 0 {
			se.params.Resolution = resolution
			se.params.Radius = se.params.Resolution * float64(se.params.Width)
		} else {
			panic("One of n, radius, resolution must be set")
		}

		if se.params.Periodic {
			se.params.Range = se.rangeInternal
		} else {
			se.params.Range = se.rangeInternal + se.params.Resolution
		}

		nfloat := float64(se.params.Width)*(se.params.Range/se.params.Radius) + 2*float64(se.padding)
		se.params.N = int(math.Ceil(nfloat))

	}

}

/*
	recalculate encoder parameters and name
*/
func (se *ScalerEncoder) recalcParams() {
	se.rangeInternal = se.params.MaxVal - se.params.MinVal

	if !se.params.Periodic {
		se.params.Resolution = se.rangeInternal/float64(se.params.N) - float64(se.params.Width)
	} else {
		se.params.Resolution = se.rangeInternal / float64(se.params.N)
	}

	se.params.Radius = float64(se.params.Width) * se.params.Resolution

	if se.params.Periodic {
		se.params.Range = se.rangeInternal
	} else {
		se.params.Range = se.rangeInternal + se.params.Resolution
	}

	se.params.Name = fmt.Sprintf("[%v:%v]", se.params.MinVal, se.params.MaxVal)

}

/* Return the bit offset of the first bit to be set in the encoder output.
For periodic encoders, this can be a negative number when the encoded output
wraps around. */
func (se *ScalerEncoder) getFirstOnBit(input float64) int {

	//if input == SENTINEL_VALUE_FOR_MISSING_DATA:
	//	return [None]
	//else:

	if input < se.params.MinVal {
		//Don't clip periodic inputs. Out-of-range input is always an error
		if se.params.ClipInput && !se.params.Periodic {

			if se.params.Verbosity > 0 {
				fmt.Printf("Clipped input %v=%d to minval %d", se.params.Name, input, se.params.MinVal)
			}
			input = se.params.MinVal
		} else {
			panic(fmt.Sprintf("Input %v less than range %v - %v", input, se.params.MinVal, se.params.MaxVal))
		}

		if se.params.Periodic {

			// Don't clip periodic inputs. Out-of-range input is always an error
			if input >= se.params.MaxVal {
				panic(fmt.Sprintf("input %v greater than periodic range %v - %v", input, se.params.MinVal, se.params.MaxVal))
			}

		} else {

			if input > se.params.MaxVal {
				if se.params.ClipInput {
					if se.params.Verbosity > 0 {
						fmt.Printf("Clipped input %v=%v to maxval %v", se.params.Name, input, se.params.MaxVal)
					}
					input = se.params.MaxVal
				} else {
					panic(fmt.Sprintf("input %v greater than range (%v - %v)", input, se.params.MinVal, se.params.MaxVal))
				}
			}
		}
	}

	centerbin := 0

	if se.params.Periodic {
		centerbin = int((input-se.params.MinVal)*float64(se.nInternal)/se.params.Range) + se.padding
	} else {
		centerbin = int(((input-se.params.MinVal)+se.params.Resolution/2)/se.params.Resolution) + se.padding
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
	if se.params.Periodic {
		bucketIdx = minbin + se.halfWidth
		if bucketIdx < 0 {
			bucketIdx += se.params.N
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
	output = make([]bool, se.params.N)
	minbin := bucketIdx
	maxbin := minbin + 2*se.halfWidth

	if se.params.Periodic {

		// Handle the edges by computing wrap-around
		if maxbin >= se.params.N {
			bottombins := maxbin - se.params.N + 1
			utils.FillSliceRangeBool(output, true, 0, bottombins)
			maxbin = se.params.N - 1
		}
		if minbin < 0 {
			topbins := -minbin
			utils.FillSliceRangeBool(output, true, se.params.N-topbins, se.params.N)
			minbin = 0
		}

	}

	if minbin < 0 {
		panic("invalid minbin")
	}
	if maxbin >= se.params.N {
		panic("invalid maxbin")
	}

	// set the output (except for periodic wraparound)
	utils.FillSliceRangeBool(output, true, minbin, maxbin+1)

	if se.params.Verbosity >= 2 {
		fmt.Println("input:", input)
		fmt.Printf("range: %v - %v \n", se.params.MinVal, se.params.MaxVal)
		fmt.Printf("n: %v width: %v resolution: %v \n", se.params.N, se.params.Width, se.params.Resolution)
		fmt.Printf("radius: %v periodic: %v \n", se.params.Radius, se.params.Periodic)
		fmt.Printf("output: %v \n", output)
	}

	//}

	return output
}
