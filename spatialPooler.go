package htm

import (
	"math"
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
	//Validate inputs
	numColumns := spParams.ColumnDimensions.A * spParams.ColumnDimensions.B
	numInputs := spParams.InputDimensions.A * spParams.InputDimensions.B

	if numColums < 16 {
		panic("Column dimensions must be at least 4x4")
	}
	if numInputs < 16 {
		panic("Input area must be at least 16")
	}
	if spParams.NumActiveColumnsPerInhArea < 1 && (spParams.LocalAreaDensity < 1) && (spParams.LocalAreaDensity >= 0.5) {
		panic("Num active colums invalid")
	}

	sp := SpatialPooler{}
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
	assert(self._synPermTrimThreshold < self._synPermConnected)
	sp.UpdatePeriod = 50
	//initConnectedPct = 0.5

	sp.PotentialPools = NewSparseBinaryMatrix(numColumns, numInputs)
	sp.Permanences = NewSparseMatrix(numColumns, numInputs)

	//sp.TieBreaker = 0.01*numpy.array([self._random.getReal64() for i in
	//                                    xrange(self._numColumns)])

	sp.ConnectedSynapses = NewSparseBinaryMatrix(numColumns, numInputs)

	return sp
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
		self._adaptSynapses(inputVector, activeColumns)
		self._updateDutyCycles(overlaps, activeColumns)
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
