package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/skelterjohn/go.matrix"
	"math"
	"math/rand"
	//"sort"
)

type ITuple struct {
	A int
	B int
}

type SpatialPooler struct {
	numColumns                 int
	numInputs                  int
	ColumnDimensions           []int
	InputDimensions            []int
	PotentialRadius            int
	PotentialPct               float64
	GlobalInhibition           bool
	NumActiveColumnsPerInhArea int
	LocalAreaDensity           float64
	StimulusThreshold          int
	SynPermInactiveDec         float64
	SynPermActiveInc           float64
	SynPermBelowStimulusInc    float64
	SynPermConnected           float64
	MinPctOverlapDutyCycles    float64
	MinPctActiveDutyCycles     float64
	DutyCyclePeriod            int
	MaxBoost                   int
	SpVerbosity                int

	// Extra parameter settings
	SynPermMin           float64
	SynPermMax           float64
	SynPermTrimThreshold float64
	UpdatePeriod         int
	InitConnectedPct     float64

	// Internal state
	Version           float64
	IterationNum      int
	IterationLearnNum int

	//random seed
	Seed int

	potentialPools *SparseBinaryMatrix
	permanences    *matrix.SparseMatrix
	tieBreaker     float64

	connectedSynapses *SparseBinaryMatrix
	//redundant
	connectedCounts []int

	overlapDutyCycles    []bool
	activeDutyCycles     []float64
	minOverlapDutyCycles []float64
	minActiveDutyCycles  []float64
	boostFactors         []bool

	inhibitionRadius int

	spVerbosity int
}

type SpParams struct {
	InputDimensions            []int
	ColumnDimensions           []int
	PotentialRadius            int
	PotentialPct               float64
	GlobalInhibition           bool
	LocalAreaDensity           float64
	NumActiveColumnsPerInhArea int
	StimulusThreshold          int
	SynPermInactiveDec         float64
	SynPermActiveInc           float64
	SynPermConnected           float64
	MinPctOverlapDutyCycle     float64
	MinPctActiveDutyCycle      float64
	DutyCyclePeriod            int
	MaxBoost                   int
	Seed                       int
	SpVerbosity                int
}

//Initializes default spatial pooler params
func NewSpParams() SpParams {
	sp := SpParams{}

	sp.InputDimensions = []int{32, 32}
	sp.ColumnDimensions = []int{64, 64}
	sp.PotentialRadius = 16
	sp.PotentialPct = 0.5
	sp.GlobalInhibition = false
	sp.LocalAreaDensity = -1.0
	sp.NumActiveColumnsPerInhArea = 10.0
	sp.StimulusThreshold = 0
	sp.SynPermInactiveDec = 0.01
	sp.SynPermActiveInc = 0.1
	sp.SynPermConnected = 0.10
	sp.MinPctOverlapDutyCycle = 0.001
	sp.MinPctActiveDutyCycle = 0.001
	sp.DutyCyclePeriod = 1000
	sp.MaxBoost = 10.0
	sp.Seed = -1
	sp.SpVerbosity = 0

	return sp
}

//Creates a new spatial pooler
func NewSpatialPooler(spParams SpParams) *SpatialPooler {
	sp := SpatialPooler{}
	//Validate inputs
	sp.numColumns = ProdInt(spParams.ColumnDimensions)
	sp.numInputs = ProdInt(spParams.InputDimensions)

	if sp.numColumns < 1 {
		panic("Must have at least 1 column")
	}
	if sp.numInputs < 1 {
		panic("must have at least 1 input")
	}
	if spParams.NumActiveColumnsPerInhArea < 1 && (spParams.LocalAreaDensity < 1) && (spParams.LocalAreaDensity >= 0.5) {
		panic("Num active colums invalid")
	}

	sp.InputDimensions = spParams.InputDimensions
	sp.ColumnDimensions = spParams.ColumnDimensions
	sp.PotentialRadius = int(mathutil.Min(spParams.PotentialRadius, sp.numInputs))
	sp.PotentialPct = spParams.PotentialPct
	sp.GlobalInhibition = spParams.GlobalInhibition
	sp.LocalAreaDensity = spParams.LocalAreaDensity
	sp.NumActiveColumnsPerInhArea = spParams.NumActiveColumnsPerInhArea
	sp.StimulusThreshold = spParams.StimulusThreshold
	sp.SynPermInactiveDec = spParams.SynPermInactiveDec
	sp.SynPermActiveInc = spParams.SynPermConnected / 10.0
	sp.SynPermConnected = spParams.SynPermConnected
	sp.MinPctOverlapDutyCycles = spParams.MinPctOverlapDutyCycle
	sp.MinPctActiveDutyCycles = spParams.MinPctActiveDutyCycle
	sp.DutyCyclePeriod = spParams.DutyCyclePeriod
	sp.MaxBoost = spParams.MaxBoost
	sp.Seed = spParams.Seed
	sp.SpVerbosity = spParams.SpVerbosity

	// Extra parameter settings
	sp.SynPermMin = 0
	sp.SynPermMax = 1
	sp.SynPermTrimThreshold = sp.SynPermActiveInc / 2.0
	if sp.SynPermTrimThreshold >= sp.SynPermConnected {
		panic("Syn perm threshold >= syn connected.")
	}
	sp.UpdatePeriod = 50
	sp.InitConnectedPct = 0.5

	/*
			# Internal state
		    version = 1.0
		    iterationNum = 0
		    iterationLearnNum = 0
	*/

	/*
			 Store the set of all inputs that are within each column's potential pool.
		     'potentialPools' is a matrix, whose rows represent cortical columns, and
		     whose columns represent the input bits. if potentialPools[i][j] == 1,
		     then input bit 'j' is in column 'i's potential pool. A column can only be
		     connected to inputs in its potential pool. The indices refer to a
		     falttenned version of both the inputs and columns. Namely, irrespective
		     of the topology of the inputs and columns, they are treated as being a
		     one dimensional array. Since a column is typically connected to only a
		     subset of the inputs, many of the entries in the matrix are 0. Therefore
		     the the potentialPool matrix is stored using the SparseBinaryMatrix
		     class, to reduce memory footprint and compuation time of algorithms that
		     require iterating over the data strcuture.
	*/
	sp.potentialPools = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)

	/*
			 Initialize the permanences for each column. Similar to the
		     'potentialPools', the permances are stored in a matrix whose rows
		     represent the cortial columns, and whose columns represent the input
		     bits. if permanences[i][j] = 0.2, then the synapse connecting
		     cortical column 'i' to input bit 'j' has a permanence of 0.2. Here we
		     also use the SparseMatrix class to reduce the memory footprint and
		     computation time of algorithms that require iterating over the data
		     structure. This permanence matrix is only allowed to have non-zero
		     elements where the potential pool is non-zero.
	*/
	//Assumes 70% sparsity
	elms := make(map[int]float64, int(float64(sp.numColumns*sp.numInputs)*0.3))
	sp.permanences = matrix.MakeSparseMatrix(elms, sp.numColumns, sp.numInputs)

	/*
			 Initialize a tiny random tie breaker. This is used to determine winning
		     columns where the overlaps are identical.
	*/
	//sp.TieBreaker = 0.01*numpy.array([random.getReal64() for i in
	//                                    xrange(numColumns)])

	/*
			 'connectedSynapses' is a similar matrix to 'permanences'
		     (rows represent cortial columns, columns represent input bits) whose
		     entries represent whether the cortial column is connected to the input
		     bit, i.e. its permanence value is greater than 'synPermConnected'. While
		     this information is readily available from the 'permanence' matrix,
		     it is stored separately for efficiency purposes.
	*/
	sp.connectedSynapses = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)

	/*
			 Stores the number of connected synapses for each column. This is simply
		     a sum of each row of 'ConnectedSynapses'. again, while this
		     information is readily available from 'ConnectedSynapses', it is
		     stored separately for efficiency purposes.
	*/
	sp.connectedCounts = make([]int, sp.numColumns)

	/*
			 Initialize the set of permanence values for each columns. Ensure that
		     each column is connected to enough input bits to allow it to be
		     activated
	*/
	for i := 0; i < sp.numColumns; i++ {
		potential := sp.mapPotential(i, true)
		sp.potentialPools.ReplaceRow(i, potential)
		perm := sp.initPermanence(potential, sp.InitConnectedPct)
		sp.updatePermanencesForColumn(perm, i, true)
	}

	sp.overlapDutyCycles = make([]bool, sp.numColumns)
	sp.activeDutyCycles = make([]float64, sp.numColumns)
	sp.minOverlapDutyCycles = make([]float64, sp.numColumns)
	sp.minActiveDutyCycles = make([]float64, sp.numColumns)
	sp.boostFactors = make([]bool, sp.numColumns)
	for i := 0; i < len(sp.boostFactors); i++ {
		sp.boostFactors[i] = true
	}

	/*
			The inhibition radius determines the size of a column's local
		    neighborhood. of a column. A cortical column must overcome the overlap
		    score of columns in his neighborhood in order to become actives. This
		    radius is updated every learning round. It grows and shrinks with the
		    average number of connected synapses per column.
	*/
	sp.inhibitionRadius = 0
	sp.updateInhibitionRadius()

	if sp.spVerbosity > 0 {
		sp.printParameters()
	}

	return &sp
}

/*
 Maps a column to its input bits. This method encapsultes the topology of
the region. It takes the index of the column as an argument and determines
what are the indices of the input vector that are located within the
column's potential pool. The return value is a list containing the indices
of the input bits. The current implementation of the base class only
supports a 1 dimensional topology of columsn with a 1 dimensional topology
of inputs. To extend this class to support 2-D topology you will need to
override this method. Examples of the expected output of this method:
* If the potentialRadius is greater than or equal to the entire input
space, (global visibility), then this method returns an array filled with
all the indices
* If the topology is one dimensional, and the potentialRadius is 5, this
method will return an array containing 5 consecutive values centered on
the index of the column (wrapping around if necessary).
* If the topology is two dimensional (not implemented), and the
potentialRadius is 5, the method should return an array containing 25
'1's, where the exact indices are to be determined by the mapping from
1-D index to 2-D position.

Parameters:
----------------------------
index: The index identifying a column in the permanence, potential
and connectivity matrices.
wrapAround: A boolean value indicating that boundaries should be
region boundaries ignored.
*/
func (sp *SpatialPooler) mapPotential(index int, wrapAround bool) []bool {
	// Distribute column over inputs uniformly
	ratio := index / mathutil.Max((sp.numColumns-1), 1)
	index = int((sp.numInputs - 1) * ratio)

	var indices []int
	indLen := 2*sp.PotentialRadius + 1

	for i := 0; i < indLen; i++ {
		temp := (i + index - sp.PotentialRadius)
		if wrapAround {
			temp = temp % sp.numInputs
		} else {
			if !(temp >= 0 && temp < sp.numInputs) {
				continue
			}
		}
		//no dupes
		exists := false
		for _, val := range indices {
			if val == temp {
				exists = true
				break
			}
		}
		if !exists {
			indices = append(indices, temp)
		}
	}

	// Select a subset of the receptive field to serve as the
	// the potential pool

	//shuffle indices
	for i := range indices {
		j := rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	sampleLen := int(float64(len(indices)) * sp.PotentialPct)
	sample := indices[:sampleLen]
	//project indices onto input mask
	mask := make([]bool, sp.numInputs)
	for i, _ := range mask {
		found := false
		for x := 0; x < len(sample); x++ {
			if sample[x] == i {
				found = true
				break
			}
		}
		mask[i] = found
	}

	return mask
}

/*
 Returns a randomly generated permanence value for a synapses that is
initialized in a connected state. The basic idea here is to initialize
permanence values very close to synPermConnected so that a small number of
learning steps could make it disconnected or connected.

Note: experimentation was done a long time ago on the best way to initialize
permanence values, but the history for this particular scheme has been lost.
*/

func (sp *SpatialPooler) initPermConnected() float64 {

	p := sp.SynPermConnected + rand.Float64()*sp.SynPermActiveInc/4.0

	// Ensure we don't have too much unnecessary precision. A full 64 bits of
	// precision causes numerical stability issues across platforms and across
	// implementations

	return float64(int(p*100000)) / 100000.0
}

/*
 Returns a randomly generated permanence value for a synapses that is to be
	initialized in a non-connected state.
*/

func (sp *SpatialPooler) initPermNonConnected() float64 {
	p := sp.SynPermConnected * rand.Float64()

	// Ensure we don't have too much unnecessary precision. A full 64 bits of
	// precision causes numerical stability issues across platforms and across
	// implementations
	return float64(int(p*100000)) / 100000.0
}

/*
 Initializes the permanences of a column. The method
returns a 1-D array the size of the input, where each entry in the
array represents the initial permanence value between the input bit
at the particular index in the array, and the column represented by
the 'index' parameter.

Parameters:
----------------------------
potential: A numpy array specifying the potential pool of the column.
Permanence values will only be generated for input bits
corresponding to indices for which the mask value is 1.
connectedPct: A value between 0 and 1 specifying the percent of the input
bits that will start off in a connected state.
*/

func (sp *SpatialPooler) initPermanence(potential []bool, connectedPct float64) []float64 {
	// Determine which inputs bits will start out as connected
	// to the inputs. Initially a subset of the input bits in a
	// column's potential pool will be connected. This number is
	// given by the parameter "connectedPct"

	perm := make([]float64, sp.numInputs)
	//var perm []int

	for i := 0; i < sp.numInputs; i++ {
		if !potential[i] {
			continue
		}
		var temp float64
		if randFloatRange(0.0, 1.0) < connectedPct {
			temp = sp.initPermConnected()
		} else {
			temp = sp.initPermNonConnected()
		}
		//Exclude low values to save memory
		if temp < sp.SynPermTrimThreshold {
			temp = 0.0
		}

		perm[i] = temp
	}

	return perm
}

/*
 This method ensures that each column has enough connections to input bits
to allow it to become active. Since a column must have at least
'stimulusThreshold' overlaps in order to be considered during the
inhibition phase, columns without such minimal number of connections, even
if all the input bits they are connected to turn on, have no chance of
obtaining the minimum threshold. For such columns, the permanence values
are increased until the minimum number of connections are formed.

Parameters:
----------------------------
perm: An array of permanence values for a column. The array is
"dense", i.e. it contains an entry for each input bit, even
if the permanence value is 0.
mask: the indices of the columns whose permanences need to be
raised.
*/

func (sp *SpatialPooler) raisePermanenceToThreshold(perm []float64, mask []int) {

	for i := 0; i < len(perm); i++ {
		if perm[i] < sp.SynPermMin {
			perm[i] = sp.SynPermMin
		} else if perm[i] > sp.SynPermMax {
			perm[i] = sp.SynPermMax
		}
	}

	for {
		numConnected := 0
		for i := 0; i < len(perm); i++ {
			if perm[i] > sp.SynPermConnected {
				numConnected++
			}
		}
		if numConnected >= sp.StimulusThreshold {
			return
		}
		for i := 0; i < len(mask); i++ {
			perm[mask[i]] += sp.SynPermBelowStimulusInc
		}

	}

}

/*
 This method updates the permanence matrix with a column's new permanence
values. The column is identified by its index, which reflects the row in
the matrix, and the permanence is given in 'dense' form, i.e. a full
arrray containing all the zeros as well as the non-zero values. It is in
charge of implementing 'clipping' - ensuring that the permanence values are
always between 0 and 1 - and 'trimming' - enforcing sparsity by zeroing out
all permanence values below '_synPermTrimThreshold'. It also maintains
the consistency between 'permanences' (the matrix storeing the
permanence values), 'connectedSynapses', (the matrix storing the bits
each column is connected to), and 'connectedCounts' (an array storing
the number of input bits each column is connected to). Every method wishing
to modify the permanence matrix should do so through this method.

Parameters:
----------------------------
perm: An array of permanence values for a column. The array is
"dense", i.e. it contains an entry for each input bit, even
if the permanence value is 0.
index: The index identifying a column in the permanence, potential
and connectivity matrices
raisePerm: a boolean value indicating whether the permanence values
should be raised until a minimum number are synapses are in
a connected state. Should be set to 'false' when a direct
assignment is required.
*/

func (sp *SpatialPooler) updatePermanencesForColumn(perm []float64, index int, raisePerm bool) {
	maskPotential := sp.potentialPools.GetRowIndices(index)
	if raisePerm {
		sp.raisePermanenceToThreshold(perm, maskPotential)
	}
	var newConnected []int
	for i := 0; i < len(perm); i++ {
		if perm[i] < sp.SynPermTrimThreshold {
			perm[i] = 0
			continue
		}

		//output[i] = perm[i] > 0
		//TODO: can be simplified if syn min/max are always 1/0
		if perm[i] < sp.SynPermMin {
			perm[i] = sp.SynPermMin
		}
		if perm[i] > sp.SynPermMax {
			perm[i] = sp.SynPermMax
		}
		if perm[i] >= sp.SynPermConnected {
			newConnected = append(newConnected, i)
		}
	}
	//TODO: replace with sparse matrix that indexes by rows
	//sp.permanences.SetRowFromDense(index, perm)
	for i := 0; i < len(perm); i++ {
		sp.permanences.Set(index, i, perm[i])
	}
	sp.connectedSynapses.ReplaceRowByIndices(index, newConnected)
	sp.connectedCounts[index] = len(newConnected)
}

//Main func, returns active array
//active arrays length is equal to # of columns
// func (sp *SpatialPooler) Compute(inputVector []bool, learn bool) []bool {
// 	if len(inputVector) != sp.numInputs {
// 		panic("input != numimputs")
// 	}

// sp.updateBookeepingVars(learn)
// //inputVector = numpy.array(inputVector, dtype=realDType)
// //inputVector.reshape(-1)

// overlaps := sp.calculateOverlap(inputVector)

// boostedOverlaps := overlaps
// // Apply boosting when learning is on
// if learn {
// 	boostedOverlaps = sp.BoostFactors * overlaps
// }

// // Apply inhibition to determine the winning columns
// activeColumns = sp.inhibitColumns(boostedOverlaps)

// if learn {
// 	adaptSynapses(inputVector, activeColumns)
// 	updateDutyCycles(overlaps, activeColumns)
// 	sp.bumpUpWeakColumns()
// 	sp.updateBoostFactors()
// 	if sp.isUpdateRound() {
// 		sp.updateInhibitionRadius()
// 		sp.updateMinDutyCycles()
// 	}

// } else {
// 	activeColumns = sp.stripNeverLearned(activeColumns)
// }

// activeArray.fill(0)
// if len(activeColumns) > 0 {
// 	activeArray[activeColumns] = 1
// }

//}

func (sp *SpatialPooler) updateBookeepingVars() {

}

func (sp *SpatialPooler) calculateOverlap(inputVector []bool) {

}

func (sp *SpatialPooler) inhibitColumns() {

}

func (sp *SpatialPooler) adaptSynapses(inputVector []bool) {

}

func (sp *SpatialPooler) updateDutyCycles() {

}

func (sp *SpatialPooler) bumpUpWeakColumns() {

}

func (sp *SpatialPooler) updateBoostFactors() {

}

func (sp *SpatialPooler) isUpdateRound() {

}

/*
The range of connectedSynapses per column, averaged for each dimension.
This vaule is used to calculate the inhibition radius. This variation of
the function supports arbitrary column dimensions.

Parameters:
----------------------------
index: The index identifying a column in the permanence, potential
and connectivity matrices.
*/

func (sp *SpatialPooler) avgConnectedSpanForColumnND(index int) float64 {
	dimensions := sp.InputDimensions

	bounds := append(dimensions[1:], 1)
	bounds = RevCumProdInt(bounds)

	connected := sp.connectedSynapses.GetRowIndices(index)
	if len(connected) == 0 {
		return 0
	}

	maxCoord := make([]int, len(dimensions))
	minCoord := make([]int, len(dimensions))
	inputMax := 0
	for i := 0; i < len(dimensions); i++ {
		if dimensions[i] > inputMax {
			inputMax = dimensions[i]
		}
	}
	for i := 0; i < len(maxCoord); i++ {
		maxCoord[i] = -1.0
		minCoord[i] = inputMax
	}
	//calc min/max of (i/bounds) % dimensions
	for _, val := range connected {
		for j := 0; j < len(dimensions); j++ {
			coord := (val / bounds[j]) % dimensions[j]
			if coord > maxCoord[j] {
				maxCoord[j] = coord
			}
			if coord < minCoord[j] {
				minCoord[j] = coord
			}
		}
	}

	sum := 0
	for i := 0; i < len(dimensions); i++ {
		sum += maxCoord[i] - minCoord[i] + 1
	}

	return float64(sum) / float64(len(dimensions))
}

/*
The average number of columns per input, taking into account the topology
of the inputs and columns. This value is used to calculate the inhibition
radius. This function supports an arbitrary number of dimensions. If the
number of column dimensions does not match the number of input dimensions,
we treat the missing, or phantom dimensions as 'ones'.
*/

func (sp *SpatialPooler) avgColumnsPerInput() float64 {

	//TODO: extend to support different number of dimensions for inputs and
	// columns
	//numDim = max(self._columnDimensions.size, self._inputDimensions.size)
	numDim := MaxInt(sp.ColumnDimensions, sp.InputDimensions)

	columnDims := sp.ColumnDimensions
	inputDims := sp.InputDimensions

	//overlay column dimensions across 1's matrix
	colDim := make([][]int, numDim[0])
	for i := 0; i < len(colDim); i++ {
		colDim[i] = make([]int, numDim[1])
		for j := 0; j < len(colDim[i]); j++ {
			if j < numDim[1] {
				colDim[i][j] = columnDims[j]
			} else {
				colDim[i][j] = 1
			}
		}
	}

	inputDim := make([][]int, numDim[0])
	for i := 0; i < len(inputDim); i++ {
		inputDim[i] = make([]int, numDim[1])
		for j := 0; j < len(inputDim[i]); j++ {
			if j < numDim[1] {
				inputDim[i][j] = inputDims[j]
			} else {
				inputDim[i][j] = 1
			}
		}
	}

	//columnsPerInput = colDim.astype(realDType) / inputDim
	sum := 0.0
	for i := 0; i < len(inputDim); i++ {
		for j := 0; j < len(inputDim[i]); j++ {
			sum += float64(colDim[i][j]) / float64(inputDim[i][j])
		}
	}

	return sum / float64(numDim[0]*numDim[1])
	//return numpy.average(columnsPerInput)
}

/*
 Update the inhibition radius. The inhibition radius is a meausre of the
square (or hypersquare) of columns that each a column is "conencted to"
on average. Since columns are are not connected to each other directly, we
determine this quantity by first figuring out how many *inputs* a column is
connected to, and then multiplying it by the total number of columns that
exist for each input. For multiple dimension the aforementioned
calculations are averaged over all dimensions of inputs and columns. This
value is meaningless if global inhibition is enabled.
*/
func (sp *SpatialPooler) updateInhibitionRadius() {

	if sp.GlobalInhibition {
		cmax := MaxIntSlice(sp.ColumnDimensions)
		sp.inhibitionRadius = cmax
		return
	}

	avgConnectedSpan := 0.0
	for i := 0; i < sp.numColumns; i++ {
		avgConnectedSpan += sp.avgConnectedSpanForColumnND(i)
	}
	avgConnectedSpan = avgConnectedSpan / float64(sp.numColumns)

	columnsPerInput := sp.avgColumnsPerInput()
	diameter := avgConnectedSpan * columnsPerInput
	radius := (diameter - 1) / 2.0
	radius = math.Max(1.0, radius)

	sp.inhibitionRadius = int(RoundPrec(radius, 0))
}

// Updates the minimum duty cycles defining normal activity for a column. A
// column with activity duty cycle below this minimum threshold is boosted.
// func (sp *SpatialPooler) updateMinDutyCycles() {
// 	if sp.GlobalInhibition || sp.inhibitionRadius > sp.numInputs {
// 		sp.updateMinDutyCyclesGlobal()
// 	} else {
// 		sp.updateMinDutyCyclesLocal()
// 	}

// }

// Updates the minimum duty cycles in a global fashion. Sets the minimum duty
// cycles for the overlap and activation of all columns to be a percent of the
// maximum in the region, specified by minPctOverlapDutyCycle and
// minPctActiveDutyCycle respectively. Functionaly it is equivalent to
// _updateMinDutyCyclesLocal, but this function exploits the globalilty of the
// compuation to perform it in a straightforward, and more efficient manner.
// func (sp *SpatialPooler) updateMinDutyCyclesGlobal() {
// 	sp.minOverlapDutyCycles.fill(sp.minPctOverlapDutyCycles * sp.overlapDutyCycles.max())
// 	sp.minActiveDutyCycles.fill(sp.minPctActiveDutyCycles * sp.activeDutyCycles.max())
// }

/*
Removes the set of columns who have never been active from the set of
active columns selected in the inhibition round. Such columns cannot
represent learned pattern and are therefore meaningless if only inference
is required.

Parameters:
----------------------------
activeColumns: An array containing the indices of the active columns
*/
func (sp *SpatialPooler) stripNeverLearned(activeColumns []int) []int {
	var result []int
	for i := 0; i < len(activeColumns); i++ {
		if sp.activeDutyCycles[activeColumns[i]] != 0 {
			result = append(result, activeColumns[i])
		}
	}

	return result
}

func (sp *SpatialPooler) printParameters() {
	fmt.Println("numInputs", sp.numInputs)
	fmt.Println("numColumns", sp.numColumns)

}

//----- Helper functions ----

func randFloatRange(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

//returns max index wise comparison
func MaxInt(a, b []int) []int {
	result := make([]int, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			result[i] = a[i]
		} else {
			result[i] = b[i]
		}
	}

	return result
}

//Returns max value from specified int slice
func MaxIntSlice(values []int) int {
	max := 0
	for i := 0; i < len(values); i++ {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}

//Returns product of set of integers
func ProdInt(vals []int) int {
	sum := 1
	for x := 0; x < len(vals); x++ {
		sum *= vals[x]
	}

	if sum == 1 {
		return 0
	} else {
		return sum
	}
}

//Returns cumulative product
func CumProdInt(vals []int) []int {
	if len(vals) < 2 {
		return vals
	}
	result := make([]int, len(vals))
	result[0] = vals[0]
	for x := 1; x < len(vals); x++ {
		result[x] = vals[x] * result[x-1]
	}

	return result
}

//Returns cumulative product starting from end
func RevCumProdInt(vals []int) []int {
	if len(vals) < 2 {
		return vals
	}
	result := make([]int, len(vals))
	result[len(vals)-1] = vals[len(vals)-1]
	for x := len(vals) - 2; x >= 0; x-- {
		result[x] = vals[x] * result[x+1]
	}

	return result
}

func RoundPrec(x float64, prec int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow * sign
}
