package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/skelterjohn/go.matrix"
	"math"
	"math/rand"
	"sort"
)

type TpOutputType int

const (
	Normal                 TpOutputType = 0
	ActiveState            TpOutputType = 1
	ActiveState1CellPerCol TpOutputType = 2
)

type TemporalPoolerParams struct {
	NumberOfCols           int
	CellsPerColumn         int
	InitialPerm            float64
	ConnectedPerm          float64
	MinThreshold           int
	NewSynapseCount        int
	PermanenceInc          float64
	PermanenceDec          float64
	PermanenceMax          int
	GlobalDecay            int
	ActivationThreshold    int
	DoPooling              bool
	SegUpdateValidDuration int
	BurnIn                 int
	CollectStats           bool
	//Seed                   int
	//verbosity=VERBOSITY,
	//checkSynapseConsistency=False, # for cpp only -- ignored
	TrivialPredictionMethods string
	PamLength                int
	MaxInfBacktrack          int
	MaxLrnBacktrack          int
	MaxAge                   int
	MaxSeqLength             int
	MaxSegmentsPerCell       int
	MaxSynapsesPerSegment    int
	outputType               TpOutputType
}

type TemporalPooler struct {
	params          TemporalPoolerParams
	numberOfCells   int
	activeColumns   []int
	cells           [][][]Segment
	lrnIterationIdx int
	iterationIdx    int
	segId           int
	CurrentOutput   []bool
	pamCounter      int
}

func NewTemportalPooler(tParams TemporalPoolerParams) *TemporalPooler {
	tp := TemporalPooler{}

	//validate args
	if tParams.PamLength <= 0 {
		panic("Pam length must be > 0")
	}

	//Fixed size CLA mode
	if maxSegmentsPerCell != -1 || maxSynapsesPerSegment != -1 {
		//validate args
		if tParams.MaxSegmentsPerCell <= 0 {
			panic("Maxsegs must be greater than 0")
		}
		if tParams.MaxSynapsesPerSegment <= 0 {
			panic("Max syns per segment must be greater than 0")
		}
		if tParams.GlobalDecay != 0.0 {
			panic("Global decay must be 0")
		}
		if tParams.MaxAge != 0 {
			panic("Max age must be 0")
		}
		if !(tParams.MaxSegmentsPerSegment >= tParams.NewSynapseCount) {
			panic("maxSynapsesPerSegment must be >= newSynapseCount")
		}

		tp.numberOfCells = tParams.NumberOfCols * tParams.CellsPerColumn

		// No point having larger expiration if we are not doing pooling
		if !tParams.DoPooling {
			tParams.SegUpdateValidDuration = 1
		}

		//Cells are indexed by column and index in the column
		// Every self.cells[column][index] contains a list of segments
		// Each segment is a structure of class Segment

		//TODO: initialize cells

		tp.lrnIterationIdx = 0
		tp.iterationIdx = 0
		tp.segID = 0

		// pamCounter gets reset to pamLength whenever we detect that the learning
		// state is making good predictions (at least half the columns predicted).
		// Whenever we do not make a good prediction, we decrement pamCounter.
		// When pamCounter reaches 0, we start the learn state over again at start
		// cells.
		tp.pamCounter = tParams.pamLength

	}

}
