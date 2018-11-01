package encoders

import (
	"fmt"
	//"github.com/cznic/mathutil"
	//"github.com/zacg/floats"
	"github.com/nupic-community/htm"
	"github.com/nupic-community/htm/utils"
	"github.com/zacg/ints"
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
	ScalerEncoderParams

	padding         int
	halfWidth       int
	rangeInternal   float64
	topDownMappingM *htm.SparseBinaryMatrix
	topDownValues   []float64
	bucketValues    []float64
	//nInternal represents the output area excluding the possible padding on each
	nInternal int
}

func NewScalerEncoder(p *ScalerEncoderParams) *ScalerEncoder {
	se := new(ScalerEncoder)
	se.ScalerEncoderParams = *p

	if se.Width%2 == 0 {
		panic("Width must be an odd number.")
	}

	se.halfWidth = (se.Width - 1) / 2

	/* For non-periodic inputs, padding is the number of bits "outside" the range,
	 on each side. I.e. the representation of minval is centered on some bit, and
	there are "padding" bits to the left of that centered bit; similarly with
	bits to the right of the center bit of maxval*/
	if !se.Periodic {
		se.padding = se.halfWidth
	}

	if se.MinVal >= se.MaxVal {
		panic("MinVal must be less than MaxVal")
	}

	se.rangeInternal = se.MaxVal - se.MinVal

	// There are three different ways of thinking about the representation. Handle
	// each case here.
	se.initEncoder(se.Width, se.MinVal, se.MaxVal, se.N,
		se.Radius, se.Resolution)

	// nInternal represents the output area excluding the possible padding on each
	// side
	se.nInternal = se.N - 2*se.padding

	// Our name
	if len(se.Name) == 0 {
		se.Name = fmt.Sprintf("[%v:%v]", se.MinVal, se.MaxVal)
	}

	if se.Width < 21 {
		fmt.Println("Number of bits in the SDR must be greater than 21")
	}

	return se
}

/*
	helper used to inititalize the encoder
*/
func (se *ScalerEncoder) initEncoder(width int, minval float64, maxval float64, n int,
	radius float64, resolution float64) {
	//handle 3 diff ways of representation

	if n != 0 {
		//crutches ;(
		if radius != 0 {
			panic("radius is not 0")
		}
		if resolution != 0 {
			panic("resolution is not 0")
		}
		if n <= width {
			panic("n less than width")
		}

		se.N = n

		//if (minval is not None and maxval is not None){

		if !se.Periodic {
			se.Resolution = se.rangeInternal / float64(se.N-se.Width)
		} else {
			se.Resolution = se.rangeInternal / float64(se.N)
		}

		se.Radius = float64(se.Width) * se.Resolution

		if se.Periodic {
			se.Range = se.rangeInternal
		} else {
			se.Range = se.rangeInternal + se.Resolution
		}

	} else { //n == 0
		if radius != 0 {
			if resolution != 0 {
				panic("resolution not 0")
			}
			se.Radius = radius
			se.Resolution = se.Radius / float64(width)
		} else if resolution != 0 {
			se.Resolution = resolution
			se.Radius = se.Resolution * float64(se.Width)
		} else {
			panic("One of n, radius, resolution must be set")
		}

		if se.Periodic {
			se.Range = se.rangeInternal
		} else {
			se.Range = se.rangeInternal + se.Resolution
		}

		nfloat := float64(se.Width)*(se.Range/se.Radius) + 2*float64(se.padding)
		se.N = int(math.Ceil(nfloat))

	}

}

/*
	recalculate encoder parameters and name
*/
func (se *ScalerEncoder) recalcParams() {
	se.rangeInternal = se.MaxVal - se.MinVal

	if !se.Periodic {
		se.Resolution = se.rangeInternal/float64(se.N) - float64(se.Width)
	} else {
		se.Resolution = se.rangeInternal / float64(se.N)
	}

	se.Radius = float64(se.Width) * se.Resolution

	if se.Periodic {
		se.Range = se.rangeInternal
	} else {
		se.Range = se.rangeInternal + se.Resolution
	}

	se.Name = fmt.Sprintf("[%v:%v]", se.MinVal, se.MaxVal)

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
		if se.ClipInput && !se.Periodic {

			if se.Verbosity > 0 {
				fmt.Printf("Clipped input %v=%v to minval %v", se.Name, input, se.MinVal)
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
				if se.ClipInput {
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
		centerbin = int(((input-se.MinVal)+se.Resolution/2)/se.Resolution) + se.padding
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
			bucketIdx += se.N
		}
	} else {
		// for non-periodic encoders, the bucket index is the index of the left bit
		bucketIdx = minbin
	}

	return []int{bucketIdx}
}

/*
 Returns encoded input
*/
func (se *ScalerEncoder) Encode(input float64, learn bool) (output []bool) {
	output = make([]bool, se.N)
	se.EncodeToSlice(input, learn, output)
	return output
}

/*
	Encodes input to specified slice. Slice should be valid length
*/
func (se *ScalerEncoder) EncodeToSlice(input float64, learn bool, output []bool) {

	// Get the bucket index to use
	bucketIdx := se.getFirstOnBit(input)

	//if len(bucketIdx) {
	//This shouldn't get hit
	//	panic("Missing input value")
	//TODO output[0:self.n] = 0 TODO: should all 1s, or random SDR be returned instead?
	//} else {
	// The bucket index is the index of the first bit to set in the output
	output = output[:se.N]

	minbin := bucketIdx
	maxbin := minbin + 2*se.halfWidth

	if se.Periodic {

		// Handle the edges by computing wrap-around
		if maxbin >= se.N {
			bottombins := maxbin - se.N + 1
			utils.FillSliceRangeBool(output, true, 0, bottombins)
			maxbin = se.N - 1
		}
		if minbin < 0 {
			topbins := -minbin
			utils.FillSliceRangeBool(output, true, se.N-topbins, (se.N - (se.N - topbins)))
			minbin = 0
		}

	}

	if minbin < 0 {
		panic("invalid minbin")
	}
	if maxbin >= se.N {
		panic("invalid maxbin")
	}

	// set the output (except for periodic wraparound)
	utils.FillSliceRangeBool(output, true, minbin, (maxbin+1)-minbin)

	if se.Verbosity >= 2 {
		fmt.Println("input:", input)
		fmt.Printf("half width:%v \n", se.Width)
		fmt.Printf("range: %v - %v \n", se.MinVal, se.MaxVal)
		fmt.Printf("n: %v width: %v resolution: %v \n", se.N, se.Width, se.Resolution)
		fmt.Printf("radius: %v periodic: %v \n", se.Radius, se.Periodic)
		fmt.Printf("output: %v \n", output)
	}

	//}

}

/*
	Return the interal topDownMappingM matrix used for handling the
	bucketInfo() and topDownCompute() methods. This is a matrix, one row per
	category (bucket) where each row contains the encoded output for that
	category.
*/
func (se *ScalerEncoder) getTopDownMapping() *htm.SparseBinaryMatrix {

	//if already calculated return
	if se.topDownMappingM != nil {
		return se.topDownMappingM
	}

	// The input scalar value corresponding to each possible output encoding
	if se.Periodic {
		se.topDownValues = make([]float64, 0, int(se.MaxVal-se.MinVal))
		start := se.MinVal + se.Resolution/2.0
		idx := 0
		for i := start; i <= se.MaxVal; i += se.Resolution {
			se.topDownValues[idx] = i
			idx++
		}
	} else {
		//Number of values is (max-min)/resolution
		se.topDownValues = make([]float64, int(math.Ceil((se.MaxVal-se.MinVal)/se.Resolution)))
		end := se.MaxVal + se.Resolution/2.0
		idx := 0
		for i := se.MinVal; i <= end; i += se.Resolution {
			se.topDownValues[idx] = i
			idx++
		}
	}

	// Each row represents an encoded output pattern
	numCategories := len(se.topDownValues)

	se.topDownMappingM = htm.NewSparseBinaryMatrix(numCategories, se.N)

	for i := 0; i < numCategories; i++ {
		value := se.topDownValues[i]
		value = math.Max(value, se.MinVal)
		value = math.Min(value, se.MaxVal)

		outputSpace := se.Encode(value, false)
		se.topDownMappingM.SetRowFromDense(i, outputSpace)
	}

	return se.topDownMappingM

}

/*
	Returns input description for bucket. Numenta implementations iface returns
	set of tuples to support diff encoder types.
*/
func (se *ScalerEncoder) getBucketInfo(buckets []int) (value float64, encoding []bool) {

	//ensure topdownmapping matrix is calculated
	se.getTopDownMapping()

	// The "category" is simply the bucket index
	category := buckets[0]
	encoding = se.topDownMappingM.GetDenseRow(category)

	if se.Periodic {
		value = (se.MinVal + (se.Resolution / 2.0) + (float64(category) * se.Resolution))
	} else {
		value = se.MinVal + (float64(category) * se.Resolution)
	}

	return value, encoding

}

/*
	Returns the value for each bucket defined by the encoder
*/
func (se *ScalerEncoder) getBucketValues() []float64 {

	if se.bucketValues == nil {
		topDownMappingM := se.getTopDownMapping()
		numBuckets := topDownMappingM.Height
		se.bucketValues = make([]float64, numBuckets)
		for i := 0; i < numBuckets; i++ {
			val, _ := se.getBucketInfo([]int{i})
			se.bucketValues[i] = val
		}
	}

	return se.bucketValues
}

/*
	top down compute
*/
func (se *ScalerEncoder) topDownCompute(encoded []bool) float64 {

	topDownMappingM := se.getTopDownMapping()

	//find "closest" match
	comps := topDownMappingM.RowAndSum(encoded)
	_, category := ints.Max(comps)

	val, _ := se.getBucketInfo([]int{category})
	return val

}

/*
	generates a text description of specified slice of ranges
*/
func (se *ScalerEncoder) generateRangeDescription(ranges []utils.TupleFloat) string {

	desc := ""
	numRanges := len(ranges)
	for idx, val := range ranges {
		if val.A == val.B {
			desc += fmt.Sprintf("%v-%v", val.A, val.B)
		} else {
			desc += fmt.Sprintf("%v", val.A)
		}
		if idx < numRanges-1 {
			desc += ","
		}
	}
	return desc

}

/*
	Decode an encoded sequence. Returns range of values
*/
func (se *ScalerEncoder) Decode(encoded []bool) []utils.TupleFloat {

	if !utils.AnyTrue(encoded) {
		return []utils.TupleFloat{}
	}

	tmpOutput := encoded[:se.N]

	// First, assume the input pool is not sampled 100%, and fill in the
	// "holes" in the encoded representation (which are likely to be present
	// if this is a coincidence that was learned by the SP).

	// Search for portions of the output that have "holes"
	maxZerosInARow := se.halfWidth

	for i := 0; i < maxZerosInARow; i++ {
		searchSeq := make([]bool, i+3)
		subLen := len(searchSeq)
		searchSeq[0] = true
		searchSeq[subLen-1] = true

		if se.Periodic {
			for j := 0; j < se.N; j++ {
				outputIndices := make([]int, subLen)

				for idx := range outputIndices {
					outputIndices[idx] = (j + idx) % se.N
				}

				if utils.BoolEq(searchSeq, utils.SubsetSliceBool(tmpOutput, outputIndices)) {
					utils.SetIdxBool(tmpOutput, outputIndices, true)
				}
			}

		} else {

			for j := 0; j < se.N-subLen+1; j++ {
				if utils.BoolEq(searchSeq, tmpOutput[j:j+subLen]) {
					utils.FillSliceRangeBool(tmpOutput, true, j, subLen)
				}
			}

		}

	}

	if se.Verbosity >= 2 {
		fmt.Println("raw output:", utils.Bool2Int(encoded[:se.N]))
		fmt.Println("filtered output:", utils.Bool2Int(tmpOutput))
	}

	// ------------------------------------------------------------------------
	// Find each run of 1's in sequence

	nz := utils.OnIndices(tmpOutput)
	//key = start index, value = run length
	runs := make([]utils.TupleInt, 0, len(nz))

	runStart := -1
	runLen := 0

	for idx, val := range tmpOutput {
		if val {
			//increment or new idx
			if runStart == -1 {
				runStart = idx
				runLen = 0
			}
			runLen++
		} else {
			if runStart != -1 {
				runs = append(runs, utils.TupleInt{runStart, runLen})
				runStart = -1
			}

		}
	}

	if runStart != -1 {
		runs = append(runs, utils.TupleInt{runStart, runLen})
		runStart = -1
	}

	// If we have a periodic encoder, merge the first and last run if they
	// both go all the way to the edges
	if se.Periodic && len(runs) > 1 {
		if runs[0].A == 0 && runs[len(runs)-1].A+runs[len(runs)-1].B == se.N {
			runs[len(runs)-1].B += runs[0].B
			runs = runs[1:]
		}
	}

	// ------------------------------------------------------------------------
	// Now, for each group of 1's, determine the "left" and "right" edges, where
	// the "left" edge is inset by halfwidth and the "right" edge is inset by
	// halfwidth.
	// For a group of width w or less, the "left" and "right" edge are both at
	// the center position of the group.

	ranges := make([]utils.TupleFloat, 0, len(runs)+2)

	for _, val := range runs {
		var left, right int
		start := val.A
		length := val.B

		if length <= se.Width {
			right = start + length/2
			left = right
		} else {
			left = start + se.halfWidth
			right = start + length - 1 - se.halfWidth
		}

		var inMin, inMax float64

		// Convert to input space.
		if !se.Periodic {
			inMin = float64(left-se.padding)*se.Resolution + se.MinVal
			inMax = float64(right-se.padding)*se.Resolution + se.MinVal
		} else {
			inMin = float64(left-se.padding)*se.Range/float64(se.nInternal) + se.MinVal
			inMax = float64(right-se.padding)*se.Range/float64(se.nInternal) + se.MinVal
		}

		// Handle wrap-around if periodic
		if se.Periodic {
			if inMin >= se.MaxVal {
				inMin -= se.Range
				inMax -= se.Range
			}
		}

		// Clip low end
		if inMin < se.MinVal {
			inMin = se.MinVal
		}
		if inMax < se.MinVal {
			inMax = se.MinVal
		}

		// If we have a periodic encoder, and the max is past the edge, break into
		// 2 separate ranges

		if se.Periodic && inMax >= se.MaxVal {
			ranges = append(ranges, utils.TupleFloat{inMin, se.MaxVal})
			ranges = append(ranges, utils.TupleFloat{se.MinVal, inMax - se.Range})
		} else {
			//clip high end
			if inMax > se.MaxVal {
				inMax = se.MaxVal
			}
			if inMin > se.MaxVal {
				inMin = se.MaxVal
			}
			ranges = append(ranges, utils.TupleFloat{inMin, inMax})
		}
	}

	//desc := se.generateRangeDescription(ranges)

	return ranges
}
