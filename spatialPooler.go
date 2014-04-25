package htm

import (
	"math"
	"math/rand"
)

type ITuple struct {
	A int
	B int
}

type SpatialPooler struct {
	NumInputs                  int
	NumColumns                 int
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

	potentialPools SparseBinaryMatrix
	permanences    SparseBinaryMatrix
	tieBreaker     float64

	connectedSynapses SparseBinaryMatrix
	//redundant
	connectedCounts []int

	overlapDutyCycles    []bool
	activeDutyCycles     []bool
	minOverlapDutyCycles []bool
	minActiveDutyCycles  []bool
	boostFactors         []bool

	inhibitionRadius int

	numColumns int
	numInputs  int
}

type SpParams struct {
	InputDimensions            ITuple
	ColumnDimensions           ITuple
	PotentialRadius            int
	PotentialPct               float64
	GlobalInhibition           bool
	LocalAreaDensity           float64
	NumActiveColumnsPerInhArea float64
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
func NewSpParams() {
	sp := SpParams{}

	sp.InputDimensions = ITuple{32, 32}
	sp.ColumnDimensions = ITuple{64, 64}
	sp.PotentialRadius = 16
	sp.PotentialPct = 0.5
	sp.GlobalInhibition = False
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
	sp.numColumns = spParams.ColumnDimensions.A * spParams.ColumnDimensions.B
	sp.numInputs = spParams.InputDimensions.A * spParams.InputDimensions.B

	if sp.numColums < 16 {
		panic("Column dimensions must be at least 4x4")
	}
	if sp.numInputs < 16 {
		panic("Input area must be at least 16")
	}
	if spParams.NumActiveColumnsPerInhArea < 1 && (spParams.LocalAreaDensity < 1) && (spParams.LocalAreaDensity >= 0.5) {
		panic("Num active colums invalid")
	}

	sp.InputDimensions = spParams.InputDimensions
	sp.ColumnDimensions = spParams.ColumnDimensions
	sp.PotentialRadius = int(math.Min(spParams.PotentialRadius, numInputs))
	sp.PotentialPct = spParams.PotentialPct
	sp.GlobalInhibition = spParams.GlobalInhibition
	sp.LocalAreaDensity = spParams.LocalAreaDensity
	sp.NumActiveColumnsPerInhArea = spParams.NumActiveColumnsPerInhArea
	sp.StimulusThreshold = spParams.StimulusThreshold
	sp.SynPermInactiveDec = spParams.SynPermInactiveDec
	sp.SynPermActiveInc = spParams.SynPermConnected / 10.0
	sp.SynPermConnected = spParams.SynPermConnected
	sp.MinPctOverlapDutyCycle = spParams.MinPctOverlapDutyCycle
	sp.MinPctActiveDutyCycle = spParams.MinPctActiveDutyCycle
	sp.DutyCyclePeriod = spParams.DutyCyclePeriod
	sp.MaxBoost = spParams.MaxBoost
	sp.Seed = spParams.Seed
	sp.SpVerbosity = spParams.SpVerbosity

	// Extra parameter settings
	sp.SynPermMin = 0.0
	sp.SynPermMax = 1.0
	sp.SynPermTrimThreshold = synPermActiveInc / 2.0
	assert(synPermTrimThreshold < synPermConnected)
	sp.UpdatePeriod = 50
	sp.initConnectedPct = 0.5

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
	sp.PotentialPools = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)

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
	sp.Permanences = NewSparseMatrix(sp.numColumns, sp.numInputs)

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
	sp.ConnectedSynapses = NewSparseBinaryMatrix(sp.numColumns, sp.numInputs)

	/*
			 Stores the number of connected synapses for each column. This is simply
		     a sum of each row of 'ConnectedSynapses'. again, while this
		     information is readily available from 'ConnectedSynapses', it is
		     stored separately for efficiency purposes.
	*/
	sp.ConnectedCounts = make([]int, sp.numColumns)

	/*
			 Initialize the set of permanence values for each columns. Ensure that
		     each column is connected to enough input bits to allow it to be
		     activated
	*/
	for i := 0; i < numColumns; i++ {
		potential := sp.mapPotential(i, true)
		sp.PotentialPools.ReplaceSparseRow(i, potential)
		perm := sp.initPermanence(potential, sp.initConnectedPct)
		sp.updatePermanencesForColumn(perm, i, true)
	}

	return sp
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
	ratio := float64(index) / max((sp.numColumns-1), 1)
	index = int((sp.numInputs - 1) * ratio)

	var indices []int
	indLen := 2*sp.PotentialRadius + 1

	for i = 0; i < indLen; i++ {
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
		for ind, val := range indices {
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

	sampleLen := int(round(len(indices) * sp.PotentialPct))
	sample := indices[:len(sampleLen)]
	//project indices onto input mask
	mask := make([]bool, sp.numInputs)
	for i, val := range mask {
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

func (sp *SpatialPooler) initPermConnected() int {

	p := sp.SynPermConnected + rand.float64()*sp.SynPermActiveInc/4.0
	//p = (synPermConnected + random.getReal64() *
	//  synPermActiveInc / 4.0)

	// Ensure we don't have too much unnecessary precision. A full 64 bits of
	// precision causes numerical stability issues across platforms and across
	// implementations
	p = int(p*100000) / 100000.0
	return p
}

/*
 Returns a randomly generated permanence value for a synapses that is to be
	initialized in a non-connected state.
*/

func (sp *SpatialPooler) initPermNonConnected() int {
	p := sp.SynPermConnected * rand.float64()

	// Ensure we don't have too much unnecessary precision. A full 64 bits of
	// precision causes numerical stability issues across platforms and across
	// implementations
	p = int(p*100000) / 100000.0
	return p
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
connectedPct: A value between 0 or 1 specifying the percent of the input
bits that will start off in a connected state.
*/

func (sp *SpatialPooler) initPermanence(potential []bool, connectedPct bool) []int {
	// Determine which inputs bits will start out as connected
	// to the inputs. Initially a subset of the input bits in a
	// column's potential pool will be connected. This number is
	// given by the parameter "connectedPct"

	perm := make([]int, sp.numInputs)
	//var perm []int

	for i := 0; i < sp.numInputs; i++ {
		if !potential[i] {
			continue
		}
		var temp int
		if rand.Float64() < connectedPct {
			temp = sp.initPermConnected()
		} else {
			temp = sp.initPermNonConnected()
		}
		//Exclude low values to save memory
		if temp < sp.SynPermTrimThreshold {
			temp = 0
		}

		perm[i] = temp
	}

	return perm
}

func (sp *SpatialPooler) raisePermanenceToThreshold(perm, maskPotential) {

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

func (sp *SpatialPooler) updatePermanencesForColumn(perm []int, index int, raisePerm bool) {
	//maskPotential :=
	maskPotential := sp.potentialPools.GetDenseRow(index)
	//maskPotential = numpy.where(potentialPools.getRow(index) > 0)[0]
	if raisePerm {
		sp.raisePermanenceToThreshold(perm, maskPotential)
	}
	var newConnected []int
	for i := 0; i < len(perm); i++ {
		if perm[i] < sp.SynPermTrimThreshold {
			perm[i] = 0
		}
		if perm[i] < sp.SynPermMin {
			perm[i] = sp.SynPermMin
		}
		if perm[i] > sp.SynPermMax {
			perm[i] = sp.SynPermMax
		}
		if perm[i] >= sp.SynPermConnected {
			newConnected = append(newConnected, perm[i])
		}
	}

	sp.permanences.SetRowFromDense(index, perm)
	sp.ConnectedSynapses.replaceSparseRow(index, newConnected)
	sp.ConnectedCounts[index] = len(newConnected)
}

//Main func, returns active array
//active arrays length is equal to # of columns
func (sp *SpatialPooler) Compute(inputVector []bool, learn bool) []bool {
	if len(inputVector) != sp.NumInputs {
		panic("input != numimputs")
	}

	sp.updateBookeepingVars(learn)
	//inputVector = numpy.array(inputVector, dtype=realDType)
	//inputVector.reshape(-1)

	overlaps := sp.calculateOverlap(inputVector)

	boostedOverlaps := overlaps
	// Apply boosting when learning is on
	if learn {
		boostedOverlaps = sp.BoostFactors * overlaps
	}

	// Apply inhibition to determine the winning columns
	activeColumns = sp.inhibitColumns(boostedOverlaps)

	if learn {
		adaptSynapses(inputVector, activeColumns)
		updateDutyCycles(overlaps, activeColumns)
		sp.bumpUpWeakColumns()
		sp.updateBoostFactors()
		if sp.isUpdateRound() {
			sp.updateInhibitionRadius()
			sp.updateMinDutyCycles()
		}

	} else {
		activeColumns = sp.stripNeverLearned(activeColumns)
	}

	activeArray.fill(0)
	if len(activeColumns) > 0 {
		activeArray[activeColumns] = 1
	}

}

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

func (sp *SpatialPooler) updateInhibitionRadius() {

}

// Updates the minimum duty cycles defining normal activity for a column. A
// column with activity duty cycle below this minimum threshold is boosted.
func (sp *SpatialPooler) updateMinDutyCycles() {
	if sp.GlobalInhibition || sp.InhibitionRadius > sp.NumInputs {
		sp.updateMinDutyCyclesGlobal()
	} else {
		sp.updateMinDutyCyclesLocal()
	}

}

// Updates the minimum duty cycles in a global fashion. Sets the minimum duty
// cycles for the overlap and activation of all columns to be a percent of the
// maximum in the region, specified by minPctOverlapDutyCycle and
// minPctActiveDutyCycle respectively. Functionaly it is equivalent to
// _updateMinDutyCyclesLocal, but this function exploits the globalilty of the
// compuation to perform it in a straightforward, and more efficient manner.
func (sp *SpatialPooler) updateMinDutyCyclesGlobal() {
	sp.minOverlapDutyCycles.fill(sp.minPctOverlapDutyCycles * sp.overlapDutyCycles.max())
	sp.minActiveDutyCycles.fill(sp.minPctActiveDutyCycles * sp.activeDutyCycles.max())
}

func (sp *SpatialPooler) stripNeverLearned(activeColumns []bool) {

}
