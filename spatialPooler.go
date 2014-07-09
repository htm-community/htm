package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/skelterjohn/go.matrix"
	"github.com/zacg/htm/utils"
	"math"
	"math/rand"
	"sort"
)

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
	MaxBoost                   float64
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

	potentialPools *DenseBinaryMatrix
	permanences    *matrix.SparseMatrix
	tieBreaker     []float64

	connectedSynapses *DenseBinaryMatrix
	//redundant
	connectedCounts []int

	overlapDutyCycles    []float64
	activeDutyCycles     []float64
	minOverlapDutyCycles []float64
	minActiveDutyCycles  []float64
	boostFactors         []float64

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
	MaxBoost                   float64
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
	sp.numColumns = utils.ProdInt(spParams.ColumnDimensions)
	sp.numInputs = utils.ProdInt(spParams.InputDimensions)

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
	sp.SynPermActiveInc = spParams.SynPermActiveInc
	sp.SynPermBelowStimulusInc = spParams.SynPermConnected / 10.0
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
	sp.potentialPools = NewDenseBinaryMatrix(sp.numColumns, sp.numInputs)

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

	sp.tieBreaker = make([]float64, sp.numColumns)
	for i := 0; i < len(sp.tieBreaker); i++ {
		sp.tieBreaker[i] = 0.01 * rand.Float64()
	}

	/*
			 'connectedSynapses' is a similar matrix to 'permanences'
		     (rows represent cortial columns, columns represent input bits) whose
		     entries represent whether the cortial column is connected to the input
		     bit, i.e. its permanence value is greater than 'synPermConnected'. While
		     this information is readily available from the 'permanence' matrix,
		     it is stored separately for efficiency purposes.
	*/
	sp.connectedSynapses = NewDenseBinaryMatrix(sp.numColumns, sp.numInputs)

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

	sp.overlapDutyCycles = make([]float64, sp.numColumns)
	sp.activeDutyCycles = make([]float64, sp.numColumns)
	sp.minOverlapDutyCycles = make([]float64, sp.numColumns)
	sp.minActiveDutyCycles = make([]float64, sp.numColumns)
	sp.boostFactors = make([]float64, sp.numColumns)
	for i := 0; i < len(sp.boostFactors); i++ {
		sp.boostFactors[i] = 1.0
	}

	/*
			The inhibition radius determines the size of a column's local
		    neighborhood. of a column. A cortical column must overcome the overlap
		    score of columns in his neighborhood in order to become actives. This
		    radius is updated every learning round. It grows and shrinks with the
		    average number of connected synapses per column.
	*/
	sp.inhibitionRadius = 0
	sp.updateInhibitionRadius(sp.avgConnectedSpanForColumnND, sp.avgColumnsPerInput)

	if sp.spVerbosity > 0 {
		sp.printParameters()
	}

	return &sp
}

//Returns number of inputs
func (sp *SpatialPooler) NumInputs() int {
	return sp.numInputs
}

//Returns number of columns
func (sp *SpatialPooler) NumColumns() int {
	return sp.numColumns
}

//Returns number of inputs
func (ssp *SpParams) NumInputs() int {
	return utils.ProdInt(ssp.InputDimensions)
}

//Returns number of columns
func (ssp *SpParams) NumColumns() int {
	return utils.ProdInt(ssp.ColumnDimensions)
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
	ratio := float64(index) / float64(mathutil.Max((sp.numColumns-1), 1))
	index = int(float64(sp.numInputs-1) * ratio)

	var indices []int
	indLen := 2*sp.PotentialRadius + 1

	for i := 0; i < indLen; i++ {
		temp := (i + index - sp.PotentialRadius)
		if wrapAround {
			temp = temp % sp.numInputs
			if temp < 0 {
				temp = sp.numInputs + temp
			}
		} else {
			if !(temp >= 0 && temp < sp.numInputs) {
				continue
			}
		}
		//no dupes
		if !utils.ContainsInt(temp, indices) {
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

	sampleLen := int(utils.RoundPrec(float64(len(indices))*sp.PotentialPct, 0))
	sample := indices[:sampleLen]
	//project indices onto input mask
	mask := make([]bool, sp.numInputs)
	for i, _ := range mask {
		mask[i] = utils.ContainsInt(i, sample)
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
		if utils.RandFloatRange(0.0, 1.0) < connectedPct {
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
all permanence values below 'synPermTrimThreshold'. It also maintains
the consistency between 'permanences' (the matrix storing the
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
		if perm[i] <= sp.SynPermTrimThreshold {
			perm[i] = 0
			continue
		}

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

/*
 This is the primary public method of the SpatialPooler class. This
 function takes a input vector and outputs the indices of the active columns.
 If 'learn' is set to True, this method also updates the permanences of the
 columns.

Parameters:
----------------------------
inputVector: a numpy array of 0's and 1's thata comprises the input to
			 the spatial pooler. The array will be treated as a one
			 dimensional array, therefore the dimensions of the array
			 do not have to much the exact dimensions specified in the
			 class constructor. In fact, even a list would suffice.
			The number of input bits in the vector must, however,
			match the number of bits specified by the call to the
			constructor. Therefore there must be a '0' or '1' in the
			array for every input bit.
learn: a boolean value indicating whether learning should be
	   performed. Learning entails updating the permanence
	   values of the synapses, and hence modifying the 'state'
	   of the model. Setting learning to 'off' freezes the SP
	   and has many uses. For example, you might want to feed in
	   various inputs and examine the resulting SDR's.
activeArray: an array whose size is equal to the number of columns.
	   Before the function returns this array will be populated
       with 1's at the indices of the active columns, and 0's
	   everywhere else.
*/
func (sp *SpatialPooler) Compute(inputVector []bool, learn bool, activeArray []bool, inhibitColumns inhibitColFunc) {
	if len(inputVector) != sp.numInputs {
		panic("input != numimputs")
	}

	sp.updateBookeepingVars(learn)
	overlaps := sp.calculateOverlap(inputVector)
	boostedOverlaps := make([]float64, len(overlaps))
	// Apply boosting when learning is on
	if learn {
		for i, val := range sp.boostFactors {
			boostedOverlaps[i] = float64(overlaps[i]) * val
		}
	}

	// Apply inhibition to determine the winning columns
	activeColumns := inhibitColumns(boostedOverlaps, sp.inhibitColumnsGlobal, sp.inhibitColumnsLocal)
	overlapsf := make([]float64, len(overlaps))
	for i, val := range overlaps {
		overlapsf[i] = float64(val)
	}

	if learn {
		sp.adaptSynapses(inputVector, activeColumns)
		sp.updateDutyCycles(overlapsf, activeColumns)
		sp.bumpUpWeakColumns()
		sp.updateBoostFactors()
		if sp.isUpdateRound() {
			sp.updateInhibitionRadius(sp.avgConnectedSpanForColumnND, sp.avgColumnsPerInput)
			sp.updateMinDutyCycles()
		}

	} else {
		activeColumns = sp.stripNeverLearned(activeColumns)
	}

	if len(activeColumns) > 0 {
		for i, _ := range activeArray {
			activeArray[i] = utils.ContainsInt(i, activeColumns)
		}
	}

}

/*
 Updates counter instance variables each round.

Parameters:
----------------------------
learn: a boolean value indicating whether learning should be
performed. Learning entails updating the permanence
values of the synapses, and hence modifying the 'state'
of the model. setting learning to 'off' might be useful
for indicating separate training vs. testing sets.
*/

func (sp *SpatialPooler) updateBookeepingVars(learn bool) {
	sp.IterationNum += 1
	if learn {
		sp.IterationLearnNum += 1
	}

}

/*
 This function determines each column's overlap with the current input
vector. The overlap of a column is the number of synapses for that column
that are connected (permance value is greater than 'synPermConnected')
to input bits which are turned on. Overlap values that are lower than
the 'stimulusThreshold' are ignored. The implementation takes advantage of
the SpraseBinaryMatrix class to perform this calculation efficiently.

Parameters:
----------------------------
inputVector: a numpy array of 0's and 1's that comprises the input to
the spatial pooler.
*/
func (sp *SpatialPooler) calculateOverlap(inputVector []bool) []int {
	overlaps := sp.connectedSynapses.RowAndSum(inputVector)
	for idx, _ := range overlaps {
		if overlaps[idx] < sp.StimulusThreshold {
			overlaps[idx] = 0
		}
	}
	return overlaps
}

func (sp *SpatialPooler) calculateOverlapPct(overlaps []int) []float64 {
	result := make([]float64, len(overlaps))
	for idx, val := range overlaps {
		result[idx] = float64(val) / float64(sp.connectedCounts[idx])
	}
	return result
}

/*
 Similar to _getNeighbors1D and _getNeighbors2D, this function Returns a
list of indices corresponding to the neighbors of a given column. Since the
permanence values are stored in such a way that information about toplogy
is lost. This method allows for reconstructing the toplogy of the inputs,
which are flattened to one array. Given a column's index, its neighbors are
defined as those columns that are 'radius' indices away from it in each
dimension. The method returns a list of the flat indices of these columns.
Parameters:
----------------------------
columnIndex: The index identifying a column in the permanence, potential
	and connectivity matrices.
dimensions: An array containg a dimensions for the column space. A 2x3
	grid will be represented by [2,3].
radius: Indicates how far away from a given column are other
	columns to be considered its neighbors. In the previous 2x3
	example, each column with coordinates:
	[2+/-radius, 3+/-radius] is considered a neighbor.
wrapAround: A boolean value indicating whether to consider columns at
	the border of a dimensions to be adjacent to columns at the
	other end of the dimension. For example, if the columns are
	layed out in one deimnsion, columns 1 and 10 will be
	considered adjacent if wrapAround is set to true:
	[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
*/

func (sp *SpatialPooler) getNeighborsND(columnIndex int, dimensions []int, radius int, wrapAround bool) []int {
	if len(dimensions) < 1 {
		panic("Dimensions empty")
	}

	bounds := append(dimensions[1:], 1)
	bounds = utils.RevCumProdInt(bounds)

	columnCoords := make([]int, len(bounds))
	for j := 0; j < len(bounds); j++ {
		columnCoords[j] = utils.Mod(columnIndex/bounds[j], dimensions[j])
	}

	rangeND := make([][]int, len(dimensions))
	for i := 0; i < len(dimensions); i++ {
		if wrapAround {
			cRange := make([]int, (radius*2)+1)
			for j := 0; j < (2*radius)+1; j++ {
				cRange[j] = utils.Mod((columnCoords[i]-radius)+j, dimensions[i])
			}
			rangeND[i] = cRange
		} else {
			var cRange []int
			for j := 0; j < (radius*2)+1; j++ {
				temp := columnCoords[i] - radius + j
				if temp >= 0 && temp < dimensions[i] {
					cRange = append(cRange, temp)
				}
			}
			rangeND[i] = cRange
		}
	}

	cp := utils.CartProductInt(rangeND)
	var neighbors []int
	for i := 0; i < len(cp); i++ {
		val := utils.DotInt(bounds, cp[i])
		if val != columnIndex && !utils.ContainsInt(val, neighbors) {
			neighbors = append(neighbors, val)
		}
	}

	return neighbors
}

/*
 Perform global inhibition. Performing global inhibition entails picking the
top 'numActive' columns with the highest overlap score in the entire
region. At most half of the columns in a local neighborhood are allowed to
be active.

Parameters:
----------------------------
overlaps: an array containing the overlap score for each column.
The overlap score for a column is defined as the number
of synapses in a "connected state" (connected synapses)
that are connected to input bits which are turned on.
density: The fraction of columns to survive inhibition.
*/

func (sp *SpatialPooler) inhibitColumnsGlobal(overlaps []float64, density float64) []int {
	//calculate num active per inhibition area
	numActive := int(density * float64(sp.numColumns))
	ov := make([]utils.TupleInt, len(overlaps))
	//TODO: if overlaps is assumed to be distinct this can be
	//  	simplified
	//a = value, b = original index
	for i := 0; i < len(ov); i++ {
		ov[i].A = int(overlaps[i])
		ov[i].B = i
	}
	//insert sort overlaps
	for i := 1; i < len(ov); i++ {
		for j := i; j > 0 && ov[j].A < ov[j-1].A; j-- {
			tmp := ov[j]
			ov[j] = ov[j-1]
			ov[j-1] = tmp
		}
	}

	result := make([]int, numActive)
	for i := 0; i < numActive; i++ {
		result[i] = ov[len(ov)-1-i].B
	}

	sort.Sort(sort.IntSlice(result))

	//return indexes of active columns
	return result
}

/*
 Performs local inhibition. Local inhibition is performed on a column by
column basis. Each column observes the overlaps of its neighbors and is
selected if its overlap score is within the top 'numActive' in its local
neighborhood. At most half of the columns in a local neighborhood are
allowed to be active.

Parameters:
----------------------------
overlaps: an array containing the overlap score for each column.
	The overlap score for a column is defined as the number
	of synapses in a "connected state" (connected synapses)
	that are connected to input bits which are turned on.
density: The fraction of columns to survive inhibition. This
	value is only an intended target. Since the surviving
	columns are picked in a local fashion, the exact fraction
	of survining columns is likely to vary.
*/

func (sp *SpatialPooler) inhibitColumnsLocal(overlaps []float64, density float64) []int {
	var activeColumns []int
	addToWinners := utils.MaxSliceFloat64(overlaps) / 1000.0

	for i := 0; i < sp.numColumns; i++ {
		mask := sp.getNeighborsND(i, sp.ColumnDimensions, sp.inhibitionRadius, false)

		ovSlice := make([]float64, len(mask))
		for idx, val := range mask {
			ovSlice[idx] = overlaps[val]
		}

		numActive := int(0.5 + density*float64(len(mask)+1))
		numBigger := 0
		for _, ov := range ovSlice {
			if ov > overlaps[i] {
				numBigger++
			}
		}

		if numBigger < numActive {
			activeColumns = append(activeColumns, i)
			overlaps[i] += addToWinners
		}
	}

	return activeColumns
}

type inhibitColumnsFunc func([]float64, float64) []int
type inhibitColFunc func(overlaps []float64, inhibitColumnsGlobal, inhibitColumnsLocal inhibitColumnsFunc) []int

/*
 Performs inhibition. This method calculates the necessary values needed to
actually perform inhibition and then delegates the task of picking the
active columns to helper functions.

Parameters:
----------------------------
overlaps: an array containing the overlap score for each column.
The overlap score for a column is defined as the number
of synapses in a "connected state" (connected synapses)
that are connected to input bits which are turned on.

*/

func (sp *SpatialPooler) InhibitColumns(overlaps []float64, inhibitColumnsGlobal, inhibitColumnsLocal inhibitColumnsFunc) []int {
	/*
			 determine how many columns should be selected in the inhibition phase.
		     This can be specified by either setting the 'numActiveColumnsPerInhArea'
		     parameter of the 'localAreaDensity' parameter when initializing the class
	*/
	density := 0.0
	if sp.LocalAreaDensity > 0 {
		density = sp.LocalAreaDensity
	} else {
		inhibitionArea := math.Pow(float64(2*sp.inhibitionRadius+1), float64(len(sp.ColumnDimensions)))
		inhibitionArea = math.Min(float64(sp.numColumns), inhibitionArea)
		density = float64(sp.NumActiveColumnsPerInhArea) / inhibitionArea
		density = math.Min(density, 0.5)
	}

	// Add our fixed little bit of random noise to the scores to help break ties.
	//overlaps += sp.tieBreaker
	for i := 0; i < len(overlaps); i++ {
		overlaps[i] += sp.tieBreaker[i]
	}

	if sp.GlobalInhibition ||
		sp.inhibitionRadius > utils.MaxSliceInt(sp.ColumnDimensions) {
		return inhibitColumnsGlobal(overlaps, density)
	} else {
		return inhibitColumnsLocal(overlaps, density)
	}

}

/*
 The primary method in charge of learning. Adapts the permanence values of
the synapses based on the input vector, and the chosen columns after
inhibition round. Permanence values are increased for synapses connected to
input bits that are turned on, and decreased for synapses connected to
inputs bits that are turned off.

Parameters:
----------------------------
inputVector: a numpy array of 0's and 1's thata comprises the input to
the spatial pooler. There exists an entry in the array
for every input bit.
activeColumns: an array containing the indices of the columns that
survived inhibition.
*/
func (sp *SpatialPooler) adaptSynapses(inputVector []bool, activeColumns []int) {
	var inputIndices []int
	for i, val := range inputVector {
		if val {
			inputIndices = append(inputIndices, i)
		}
	}

	permChanges := make([]float64, sp.numInputs)
	utils.FillSliceFloat64(permChanges, -1*sp.SynPermInactiveDec)
	for _, val := range inputIndices {
		permChanges[val] = sp.SynPermActiveInc
	}

	for _, ac := range activeColumns {
		perm := make([]float64, sp.numInputs)
		mask := sp.potentialPools.GetRowIndices(ac)
		for j := 0; j < sp.numInputs; j++ {
			if utils.ContainsInt(j, mask) {
				perm[j] = permChanges[j] + sp.permanences.Get(ac, j)
			} else {
				perm[j] = sp.permanences.Get(ac, j)
			}

		}
		sp.updatePermanencesForColumn(perm, ac, true)
	}

}

/*
 Updates the duty cycles for each column. The OVERLAP duty cycle is a moving
average of the number of inputs which overlapped with the each column. The
ACTIVITY duty cycles is a moving average of the frequency of activation for
each column.

Parameters:
----------------------------
overlaps: an array containing the overlap score for each column.
The overlap score for a column is defined as the number
of synapses in a "connected state" (connected synapses)
that are connected to input bits which are turned on.
activeColumns: An array containing the indices of the active columns,
the sprase set of columns which survived inhibition
*/
func (sp *SpatialPooler) updateDutyCycles(overlaps []float64, activeColumns []int) {
	overlapArray := make([]int, sp.numColumns)
	activeArray := make([]int, sp.numColumns)

	for i, val := range overlaps {
		if val > 0 {
			overlapArray[i] = 1
		}
	}

	if len(activeColumns) > 0 {
		for _, val := range activeColumns {
			activeArray[val] = 1
		}
	}

	period := sp.DutyCyclePeriod
	if period > sp.IterationNum {
		period = sp.IterationNum
	}

	sp.overlapDutyCycles = updateDutyCyclesHelper(
		sp.overlapDutyCycles,
		overlapArray,
		period,
	)

	sp.activeDutyCycles = updateDutyCyclesHelper(
		sp.activeDutyCycles,
		activeArray,
		period,
	)
}

/*
 This method increases the permanence values of synapses of columns whose
activity level has been too low. Such columns are identified by having an
overlap duty cycle that drops too much below those of their peers. The
permanence values for such columns are increased.
*/
func (sp *SpatialPooler) bumpUpWeakColumns() {
	var weakColumns []int
	for i, val := range sp.overlapDutyCycles {
		if val < sp.minOverlapDutyCycles[i] {
			weakColumns = append(weakColumns, i)
		}
	}

	for _, col := range weakColumns {
		perm := make([]float64, sp.numInputs)
		for j := 0; j < sp.numInputs; j++ {
			perm[j] = sp.permanences.Get(col, j)
		}

		maskPotential := sp.potentialPools.GetRowIndices(col)
		for _, mpot := range maskPotential {
			perm[mpot] += sp.SynPermBelowStimulusInc
		}
		sp.updatePermanencesForColumn(perm, col, false)
	}

}

/*
 Update the boost factors for all columns. The boost factors are used to
increase the overlap of inactive columns to improve their chances of
becoming active. and hence encourage participation of more columns in the
learning process. This is a line defined as: y = mx + b boost =
(1-maxBoost)/minDuty * dutyCycle + maxFiringBoost. Intuitively this means
that columns that have been active enough have a boost factor of 1, meaning
their overlap is not boosted. Columns whose active duty cycle drops too much
below that of their neighbors are boosted depending on how infrequently they
have been active. The more infrequent, the more they are boosted. The exact
boost factor is linearly interpolated between the points (dutyCycle:0,
boost:maxFiringBoost) and (dutyCycle:minDuty, boost:1.0).

boostFactor
^
maxBoost _ |
|\
| \
1 _ | \ _ _ _ _ _ _ _
|
+--------------------> activeDutyCycle
|
minActiveDutyCycle
*/

func (sp *SpatialPooler) updateBoostFactors() {
	for i, val := range sp.minActiveDutyCycles {
		if val > 0 {
			sp.boostFactors[i] = ((1.0 - sp.MaxBoost) /
				sp.minActiveDutyCycles[i] * sp.activeDutyCycles[i]) + sp.MaxBoost
		}
	}

	for i, val := range sp.activeDutyCycles {
		if val > sp.minActiveDutyCycles[i] {
			sp.boostFactors[i] = 1.0
		}
	}

}

/*
 returns true if the enough rounds have passed to warrant updates of
 duty cycles
*/
func (sp *SpatialPooler) isUpdateRound() bool {
	return (sp.IterationNum % sp.UpdatePeriod) == 0
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
	bounds = utils.RevCumProdInt(bounds)

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
	numDim := mathutil.Max(len(sp.ColumnDimensions), len(sp.InputDimensions))
	columnDims := sp.ColumnDimensions
	inputDims := sp.InputDimensions

	//overlay column dimensions across 1's matrix
	colDim := make([]int, numDim)
	inputDim := make([]int, numDim)

	for i := 0; i < numDim; i++ {
		if i < len(columnDims) {
			colDim[i] = columnDims[i]
		} else {
			colDim[i] = 1
		}

		if i < numDim {
			inputDim[i] = inputDims[i]
		} else {
			inputDim[i] = 1
		}

	}

	sum := 0.0
	for i := 0; i < len(inputDim); i++ {
		sum += float64(colDim[i]) / float64(inputDim[i])
	}
	return sum / float64(numDim)
}

type avgConnectedSpanForColumnNDFunc func(int) float64
type avgColumnsPerInputFunc func() float64

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
func (sp *SpatialPooler) updateInhibitionRadius(avgConnectedSpanForColumnND avgConnectedSpanForColumnNDFunc,
	avgColumnsPerInput avgColumnsPerInputFunc) {

	if sp.GlobalInhibition {
		cmax := utils.MaxSliceInt(sp.ColumnDimensions)
		sp.inhibitionRadius = cmax
		return
	}

	avgConnectedSpan := 0.0
	for i := 0; i < sp.numColumns; i++ {
		avgConnectedSpan += avgConnectedSpanForColumnND(i)
	}
	avgConnectedSpan = avgConnectedSpan / float64(sp.numColumns)

	columnsPerInput := avgColumnsPerInput()
	diameter := avgConnectedSpan * columnsPerInput
	radius := (diameter - 1) / 2.0
	radius = math.Max(1.0, radius)

	sp.inhibitionRadius = int(utils.RoundPrec(radius, 0))
}

/*
Updates the minimum duty cycles defining normal activity for a column. A
column with activity duty cycle below this minimum threshold is boosted.
*/
func (sp *SpatialPooler) updateMinDutyCycles() {
	if sp.GlobalInhibition || sp.inhibitionRadius > sp.numInputs {
		sp.updateMinDutyCyclesGlobal()
	} else {
		sp.updateMinDutyCyclesLocal(sp.getNeighborsND)
	}

}

/*
Updates the minimum duty cycles in a global fashion. Sets the minimum duty
cycles for the overlap and activation of all columns to be a percent of the
maximum in the region, specified by minPctOverlapDutyCycle and
minPctActiveDutyCycle respectively. Functionaly it is equivalent to
updateMinDutyCyclesLocal, but this function exploits the globalilty of the
compuation to perform it in a straightforward, and more efficient manner.
*/
func (sp *SpatialPooler) updateMinDutyCyclesGlobal() {
	minOverlap := sp.MinPctOverlapDutyCycles * utils.MaxSliceFloat64(sp.overlapDutyCycles)
	utils.FillSliceFloat64(sp.minOverlapDutyCycles, minOverlap)
	minActive := sp.MinPctActiveDutyCycles * utils.MaxSliceFloat64(sp.activeDutyCycles)
	utils.FillSliceFloat64(sp.minActiveDutyCycles, minActive)
}

type getNeighborsNDFunc func(int, []int, int, bool) []int

/*
 Updates the minimum duty cycles. The minimum duty cycles are determined
locally. Each column's minimum duty cycles are set to be a percent of the
maximum duty cycles in the column's neighborhood. Unlike
updateMinDutyCyclesGlobal, here the values can be quite different for
different columns.
*/
func (sp *SpatialPooler) updateMinDutyCyclesLocal(getNeighborsND getNeighborsNDFunc) {

	for i := 0; i < sp.numColumns; i++ {
		maskNeighbors := getNeighborsND(i, sp.ColumnDimensions, sp.inhibitionRadius, false)
		maskNeighbors = append(maskNeighbors, i)

		maxOverlap := utils.MaxSliceFloat64(utils.SubsetSliceFloat64(sp.overlapDutyCycles, maskNeighbors))
		sp.minOverlapDutyCycles[i] = maxOverlap * sp.MinPctOverlapDutyCycles

		maxActive := utils.MaxSliceFloat64(utils.SubsetSliceFloat64(sp.activeDutyCycles, maskNeighbors))
		sp.minActiveDutyCycles[i] = maxActive * sp.MinPctActiveDutyCycles
	}

}

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

/*
 Updates a duty cycle estimate with a new value. This is a helper
function that is used to update several duty cycle variables in
the Column class, such as: overlapDutyCucle, activeDutyCycle,
minPctDutyCycleBeforeInh, minPctDutyCycleAfterInh, etc. returns
the updated duty cycle. Duty cycles are updated according to the following
formula:

			(period - 1)*dutyCycle + newValue
dutyCycle := ----------------------------------
						period

Parameters:
----------------------------
dutyCycles: An array containing one or more duty cycle values that need
to be updated
newInput: A new numerical value used to update the duty cycle
period: The period of the duty cycle
*/
func updateDutyCyclesHelper(dutyCycles []float64, newInput []int, period int) []float64 {
	if period < 1.0 {
		panic("period can't be less than 1")
	}
	pf := float64(period)
	result := make([]float64, len(dutyCycles))
	for i, val := range dutyCycles {
		result[i] = (val*(pf-1.0) + float64(newInput[i])) / pf
	}
	return result
}
