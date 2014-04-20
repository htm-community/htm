package htm

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

	//# Extra parameter settings
	SynPermMin           float64
	SynPermMax           float64
	SynPermTrimThreshold float64
	UpdatePeriod         int
	InitConnectedPct     float64
}

func (sp *SpatialPooler) NewSpatialPooler() {

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
	activeColumns = self._inhibitColumns(boostedOverlaps)

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

func (sp *SpatialPooler) updateMinDutyCycles() {

}

func (sp *SpatialPooler) stripNeverLearned(activeColumns []bool) {

}
