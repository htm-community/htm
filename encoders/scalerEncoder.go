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
	params          *ScalerEncoderParams
	padding         int
	halfWidth       int
	rangeInternal   float64
	topDownMappingM *htm.SparseBinaryMatrix
	topDownValues   []float64
	bucketValues    []float64
	//nInternal represents the output area excluding the possible padding on each
	nInternal int
}

func NewScalerEncoder(params *ScalerEncoderParams) *ScalerEncoder {
	se := new(ScalerEncoder)
	se.params = params

	if params.Width%2 == 0 {
		panic("Width must be an odd number.")
	}

	se.halfWidth = (params.Width - 1) / 2

	/* For non-periodic inputs, padding is the number of bits "outside" the range,
	 on each side. I.e. the representation of minval is centered on some bit, and
	there are "padding" bits to the left of that centered bit; similarly with
	bits to the right of the center bit of maxval*/
	if !params.Periodic {
		se.padding = se.halfWidth
	}

	if params.MinVal >= params.MaxVal {
		panic("MinVal must be less than MaxVal")
	}

	se.rangeInternal = params.MaxVal - params.MinVal

	// There are three different ways of thinking about the representation. Handle
	// each case here.
	se.initEncoder(params.Width, params.MinVal, params.MaxVal, params.N,
		params.Radius, params.Resolution)

	// nInternal represents the output area excluding the possible padding on each
	// side
	se.nInternal = params.N - 2*se.padding

	// Our name
	if len(params.Name) == 0 {
		params.Name = fmt.Sprintf("[%v:%v]", params.MinVal, params.MaxVal)
	}

	if params.Width < 21 {
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
			utils.FillSliceRangeBool(output, true, se.params.N-topbins, (se.params.N - (se.params.N - topbins)))
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
	utils.FillSliceRangeBool(output, true, minbin, (maxbin+1)-minbin)

	if se.params.Verbosity >= 2 {
		fmt.Println("input:", input)
		fmt.Printf("half width:%v \n", se.params.Width)
		fmt.Printf("range: %v - %v \n", se.params.MinVal, se.params.MaxVal)
		fmt.Printf("n: %v width: %v resolution: %v \n", se.params.N, se.params.Width, se.params.Resolution)
		fmt.Printf("radius: %v periodic: %v \n", se.params.Radius, se.params.Periodic)
		fmt.Printf("output: %v \n", output)
	}

	//}

	return output
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
	if se.params.Periodic {
		se.topDownValues = make([]float64, 0, int(se.params.MaxVal-se.params.MinVal))
		start := se.params.MinVal + se.params.Resolution/2.0
		idx := 0
		for i := start; i <= se.params.MaxVal; i += se.params.Resolution {
			se.topDownValues[idx] = i
			idx++
		}
	} else {
		//Number of values is (max-min)/resolution
		se.topDownValues = make([]float64, int(math.Ceil((se.params.MaxVal-se.params.MinVal)/se.params.Resolution)))
		end := se.params.MaxVal + se.params.Resolution/2.0
		idx := 0
		for i := se.params.MinVal; i <= end; i += se.params.Resolution {
			se.topDownValues[idx] = i
			idx++
		}
	}

	// Each row represents an encoded output pattern
	numCategories := len(se.topDownValues)

	se.topDownMappingM = htm.NewSparseBinaryMatrix(numCategories, se.params.N)

	for i := 0; i < numCategories; i++ {
		value := se.topDownValues[i]
		value = math.Max(value, se.params.MinVal)
		value = math.Min(value, se.params.MaxVal)

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

	if se.params.Periodic {
		value = (se.params.MinVal + (se.params.Resolution / 2.0) + (float64(category) * se.params.Resolution))
	} else {
		value = se.params.MinVal + (float64(category) * se.params.Resolution)
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

	tmpOutput := encoded[:se.params.N]

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

		if se.params.Periodic {
			for j := 0; j < se.params.N; j++ {
				outputIndices := make([]int, subLen)

				for idx, _ := range outputIndices {
					outputIndices[idx] = (j + idx) % se.params.N
				}

				if utils.BoolEq(searchSeq, utils.SubsetSliceBool(tmpOutput, outputIndices)) {
					utils.SetIdxBool(tmpOutput, outputIndices, true)
				}
			}

		} else {

			for j := 0; j < se.params.N-subLen+1; j++ {
				if utils.BoolEq(searchSeq, tmpOutput[j:j+subLen]) {
					utils.FillSliceRangeBool(tmpOutput, true, j, subLen)
				}
			}

		}

	}

	if se.params.Verbosity >= 2 {
		fmt.Println("raw output:", utils.Bool2Int(encoded[:se.params.N]))
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
	if se.params.Periodic && len(runs) > 1 {
		if runs[0].A == 0 && runs[len(runs)-1].A+runs[len(runs)-1].B == se.params.N {
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

		if length <= se.params.Width {
			right = start + length/2
			left = right
		} else {
			left = start + se.halfWidth
			right = start + length - 1 - se.halfWidth
		}

		var inMin, inMax float64

		// Convert to input space.
		if !se.params.Periodic {
			inMin = float64(left-se.padding)*se.params.Resolution + se.params.MinVal
			inMax = float64(right-se.padding)*se.params.Resolution + se.params.MinVal
		} else {
			inMin = float64(left-se.padding)*se.params.Range/float64(se.nInternal) + se.params.MinVal
			inMax = float64(right-se.padding)*se.params.Range/float64(se.nInternal) + se.params.MinVal
		}

		// Handle wrap-around if periodic
		if se.params.Periodic {
			if inMin >= se.params.MaxVal {
				inMin -= se.params.Range
				inMax -= se.params.Range
			}
		}

		// Clip low end
		if inMin < se.params.MinVal {
			inMin = se.params.MinVal
		}
		if inMax < se.params.MinVal {
			inMax = se.params.MinVal
		}

		// If we have a periodic encoder, and the max is past the edge, break into
		// 2 separate ranges

		if se.params.Periodic && inMax >= se.params.MaxVal {
			ranges = append(ranges, utils.TupleFloat{inMin, se.params.MaxVal})
			ranges = append(ranges, utils.TupleFloat{se.params.MinVal, inMax - se.params.Range})
		} else {
			//clip high end
			if inMax > se.params.MaxVal {
				inMax = se.params.MaxVal
			}
			if inMin > se.params.MaxVal {
				inMin = se.params.MaxVal
			}
			ranges = append(ranges, utils.TupleFloat{inMin, inMax})
		}
	}

	//desc := se.generateRangeDescription(ranges)

	return ranges
}
