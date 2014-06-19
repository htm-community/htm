package htm

import (
	//"fmt"
	//"github.com/cznic/mathutil"
	"github.com/zacg/floats"
	"github.com/zacg/go.matrix"
	//"math"
	//"math/rand"
	//"sort"
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
	PermanenceMax          float64
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

type DynamicState struct {
	//orginally dynamic vars
	lrnActiveState     *SparseBinaryMatrix // t
	lrnActiveStateLast *SparseBinaryMatrix // t-1

	lrnPredictedState     *SparseBinaryMatrix
	lrnPredictedStateLast *SparseBinaryMatrix

	infActiveState          *SparseBinaryMatrix
	infActiveStateLast      *SparseBinaryMatrix
	infActiveStateBackup    *SparseBinaryMatrix
	infActiveStateCandidate *SparseBinaryMatrix

	infPredictedState          *SparseBinaryMatrix
	infPredictedStateLast      *SparseBinaryMatrix
	infPredictedStateBackup    *SparseBinaryMatrix
	infPredictedStateCandidate *SparseBinaryMatrix

	cellConfidence          *matrix.DenseMatrix
	cellConfidenceLast      *matrix.DenseMatrix
	cellConfidenceCandidate *matrix.DenseMatrix

	colConfidence          []float64
	colConfidenceLast      []float64
	colConfidenceCandidate []float64
}

func (ds *DynamicState) Copy() *DynamicState {
	result := new(DynamicState)
	result.lrnActiveState = ds.lrnActiveState.Copy()
	result.lrnActiveStateLast = ds.lrnActiveStateLast.Copy()

	result.lrnPredictedState = ds.lrnPredictedState.Copy()
	result.lrnPredictedStateLast = ds.lrnPredictedStateLast.Copy()

	result.infActiveState = ds.infActiveState.Copy()
	result.infActiveStateLast = ds.infActiveStateLast.Copy()
	result.infActiveStateBackup = ds.infActiveStateBackup.Copy()
	result.infActiveStateCandidate = ds.infActiveStateCandidate.Copy()

	result.infPredictedState = ds.infPredictedState.Copy()
	result.infPredictedStateLast = ds.infPredictedStateLast.Copy()
	result.infPredictedStateBackup = ds.infPredictedStateBackup.Copy()
	result.infPredictedStateCandidate = ds.infPredictedStateCandidate.Copy()

	result.cellConfidence = ds.cellConfidence.Copy()
	result.cellConfidenceCandidate = ds.cellConfidenceCandidate.Copy()
	result.cellConfidenceLast = ds.cellConfidenceLast.Copy()

	copy(result.colConfidence, ds.colConfidence)
	copy(result.colConfidenceCandidate, ds.colConfidenceCandidate)
	copy(result.colConfidenceLast, ds.colConfidenceLast)

	return result
}

type TemporalPooler struct {
	params              TemporalPoolerParams
	numberOfCells       int
	activeColumns       []int
	cells               [][][]Segment
	lrnIterationIdx     int
	iterationIdx        int
	segId               int
	CurrentOutput       *SparseBinaryMatrix
	pamCounter          int
	avgInputDensity     float64
	avgLearnedSeqLength float64

	//ephemeral state
	segmentUpdates map[TupleInt][]UpdateState
	/*
	 	 NOTE: We don't use the same backtrack buffer for inference and learning
	     because learning has a different metric for determining if an input from
	     the past is potentially useful again for backtracking.

	     Our inference backtrack buffer. This keeps track of up to
	     maxInfBacktrack of previous input. Each entry is a list of active column
	     inputs.
	*/
	prevInfPatterns [][]int

	/*
			 Our learning backtrack buffer. This keeps track of up to maxLrnBacktrack
		     of previous input. Each entry is a list of active column inputs
	*/

	prevLrnPatterns [][]int

	DynamicState *DynamicState
}

func NewTemportalPooler(tParams TemporalPoolerParams) *TemporalPooler {
	tp := new(TemporalPooler)

	//validate args
	if tParams.PamLength <= 0 {
		panic("Pam length must be > 0")
	}

	//Fixed size CLA mode
	if tParams.MaxSegmentsPerCell != -1 || tParams.MaxSynapsesPerSegment != -1 {
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
		if !(tParams.MaxSynapsesPerSegment >= tParams.NewSynapseCount) {
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
		tp.segId = 0

		// pamCounter gets reset to pamLength whenever we detect that the learning
		// state is making good predictions (at least half the columns predicted).
		// Whenever we do not make a good prediction, we decrement pamCounter.
		// When pamCounter reaches 0, we start the learn state over again at start
		// cells.
		tp.pamCounter = tParams.PamLength

	}

	return tp
}

//Returns new segId
func (su *TemporalPooler) GetSegId() int {
	result := su.segId
	su.segId++
	return result
}

/*
	 Compute the column confidences given the cell confidences. If
	None is passed in for cellConfidences, it uses the stored cell confidences
	from the last compute.

	param cellConfidences Cell confidences to use, or None to use the
	the current cell confidences.

	returns Column confidence scores
*/

func (su *TemporalPooler) columnConfidences() []float64 {
	//ignore cellconfidence param for now
	return su.DynamicState.colConfidence
}

/*
 Top-down compute - generate expected input given output of the TP
	param topDownIn top down input from the level above us
	returns best estimate of the TP input that would have generated bottomUpOut.
*/

func (su *TemporalPooler) topDownCompute() []float64 {
	/*
			 For now, we will assume there is no one above us and that bottomUpOut is
		     simply the output that corresponds to our currently stored column
		     confidences.

		     Simply return the column confidences
	*/

	return su.columnConfidences()
}

/*
 This function gives the future predictions for <nSteps> timesteps starting
from the current TP state. The TP is returned to its original state at the
end before returning.

- We save the TP state.
- Loop for nSteps
- Turn-on with lateral support from the current active cells
- Set the predicted cells as the next step's active cells. This step
in learn and infer methods use input here to correct the predictions.
We don't use any input here.
- Revert back the TP state to the time before prediction

param nSteps The number of future time steps to be predicted
returns all the future predictions - a numpy array of type "float32" and
shape (nSteps, numberOfCols).
The ith row gives the tp prediction for each column at
a future timestep (t+i+1).
*/

func (tp *TemporalPooler) predict(nSteps int) *matrix.DenseMatrix {
	// Save the TP dynamic state, we will use to revert back in the end
	pristineTPDynamicState := tp.DynamicState.Copy()

	if nSteps <= 0 {
		panic("nSteps must be greater than zero")
	}

	// multiStepColumnPredictions holds all the future prediction.
	var elements []float64
	multiStepColumnPredictions := matrix.MakeDenseMatrix(elements, nSteps, tp.params.NumberOfCols)

	// This is a (nSteps-1)+half loop. Phase 2 in both learn and infer methods
	// already predicts for timestep (t+1). We use that prediction for free and
	// save the half-a-loop of work.

	step := 0
	for {
		multiStepColumnPredictions.FillRow(step, tp.topDownCompute())
		if step == nSteps-1 {
			break
		}
		step += 1

		//Copy t-1 into t
		tp.DynamicState.infActiveState = tp.DynamicState.infActiveStateLast
		tp.DynamicState.infPredictedState = tp.DynamicState.infPredictedStateLast
		tp.DynamicState.cellConfidence = tp.DynamicState.cellConfidenceLast

		// Predicted state at "t-1" becomes the active state at "t"
		tp.DynamicState.infActiveState = tp.DynamicState.infPredictedState

		// Predicted state and confidence are set in phase2.
		tp.DynamicState.infPredictedState.Clear()
		tp.DynamicState.cellConfidence.Fill(0.0)
		tp.inferPhase2()
	}

	// Revert the dynamic state to the saved state
	tp.DynamicState = pristineTPDynamicState

	return multiStepColumnPredictions

}

/*
 This routine computes the activity level of a segment given activeState.
It can tally up only connected synapses (permanence >= connectedPerm), or
all the synapses of the segment, at either t or t-1.
*/

func (tp *TemporalPooler) getSegmentActivityLevel(seg Segment, activeState *SparseBinaryMatrix, connectedSynapsesOnly bool) int {
	activity := 0
	if connectedSynapsesOnly {
		for _, val := range seg.syns {
			if val.Permanence >= tp.params.ConnectedPerm {
				if activeState.Get(val.SrcCellIdx, val.SrcCellCol) {
					activity++
				}
			}
		}
	} else {
		for _, val := range seg.syns {
			if activeState.Get(val.SrcCellIdx, val.SrcCellCol) {
				activity++
			}
		}
	}

	return activity
}

/*
	 A segment is active if it has >= activationThreshold connected
	synapses that are active due to activeState.
*/

func (tp *TemporalPooler) isSegmentActive(seg Segment, activeState *SparseBinaryMatrix) bool {

	if len(seg.syns) < tp.params.ActivationThreshold {
		return false
	}

	activity := 0
	for _, val := range seg.syns {
		if val.Permanence >= tp.params.ConnectedPerm {
			if activeState.Get(val.SrcCellIdx, val.SrcCellCol) {
				activity++
				if activity >= tp.params.ActivationThreshold {
					return true
				}
			}

		}
	}

	return false
}

/*
 Phase 2 for the inference state. The computes the predicted state, then
checks to insure that the predicted state is not over-saturated, i.e.
look too close like a burst. This indicates that there were so many
separate paths learned from the current input columns to the predicted
input columns that bursting on the current input columns is most likely
generated mix and match errors on cells in the predicted columns. If
we detect this situation, we instead turn on only the start cells in the
current active columns and re-generate the predicted state from those.

returns True if we have a decent guess as to the next input.
Returing False from here indicates to the caller that we have
reached the end of a learned sequence.

This looks at:
- infActiveState

This modifies:
-  infPredictedState
-  colConfidence
-  cellConfidence
*/

func (tp *TemporalPooler) inferPhase2() bool {
	// Init to zeros to start
	tp.DynamicState.infPredictedState.Clear()
	tp.DynamicState.cellConfidence.Fill(0)
	FillSliceFloat64(tp.DynamicState.colConfidence, 0)

	// Phase 2 - Compute new predicted state and update cell and column
	// confidences
	for c := 0; c < tp.params.NumberOfCols; c++ {
		for i := 0; i < tp.params.CellsPerColumn; i++ {
			// For each segment in the cell
			for _, seg := range tp.cells[c][i] {
				// Check if it has the min number of active synapses
				numActiveSyns := tp.getSegmentActivityLevel(seg, tp.DynamicState.infActiveState, false)
				if numActiveSyns < tp.params.ActivationThreshold {
					continue
				}

				//Incorporate the confidence into the owner cell and column
				dc := seg.dutyCycle(false, false)
				tp.DynamicState.cellConfidence.Set(c, i, tp.DynamicState.cellConfidence.Get(c, i)+dc)
				tp.DynamicState.colConfidence[c] += dc

				if tp.isSegmentActive(seg, tp.DynamicState.infActiveState) {
					tp.DynamicState.infPredictedState.Set(c, i, true)
				}
			}
		}

	}

	// Normalize column and cell confidences
	sumConfidences := SumSliceFloat64(tp.DynamicState.colConfidence)

	if sumConfidences > 0 {
		floats.DivConst(sumConfidences, tp.DynamicState.colConfidence)
		tp.DynamicState.cellConfidence.DivScaler(sumConfidences)
	}

	// Are we predicting the required minimum number of columns?
	numPredictedCols := float64(tp.DynamicState.infPredictedState.TotalTrueCols())

	return numPredictedCols >= (0.5 * tp.avgInputDensity)

}

/*
Computes output for both learning and inference. In both cases, the
output is the boolean OR of activeState and predictedState at t.
Stores currentOutput for checkPrediction.
*/

func (tp *TemporalPooler) computeOutput() []bool {

	switch tp.params.outputType {
	case ActiveState1CellPerCol:
		// Fire only the most confident cell in columns that have 2 or more
		// active cells

		mostActiveCellPerCol := tp.DynamicState.cellConfidence.ArgMaxCols()
		tp.CurrentOutput = NewSparseBinaryMatrix(tp.DynamicState.infActiveState.Height, tp.DynamicState.infActiveState.Width)

		// Turn on the most confident cell in each column. Note here that
		// Columns refers to TP columns, even though each TP column is a row
		// in the matrix.
		for i := 0; i < tp.CurrentOutput.Height; i++ {
			//only on active cols
			if len(tp.DynamicState.infActiveState.GetRowIndices(i)) != 0 {
				tp.CurrentOutput.Set(i, mostActiveCellPerCol[i], true)
			}
		}

		break
	case ActiveState:
		tp.CurrentOutput = tp.DynamicState.infActiveState.Copy()
		break
	case Normal:
		tp.CurrentOutput = tp.DynamicState.infPredictedState.Or(tp.DynamicState.infActiveState)
		break
	default:
		panic("Unknown output type")
	}

	return tp.CurrentOutput.Flatten()
}

/*
Update our moving average of learned sequence length.
*/

func (tp *TemporalPooler) updateAvgLearnedSeqLength(prevSeqLength float64) {
	alpha := 0.0
	if tp.lrnIterationIdx < 100 {
		alpha = 0.5
	} else {
		alpha = 0.1
	}

	tp.avgLearnedSeqLength = ((1.0-alpha)*tp.avgLearnedSeqLength + (alpha * prevSeqLength))
}

/*
 Update the inference active state from the last set of predictions
and the current bottom-up.

This looks at:
- infPredictedState['t-1']
This modifies:
- infActiveState['t']

param activeColumns list of active bottom-ups
param useStartCells If true, ignore previous predictions and simply turn on
the start cells in the active columns
returns True if the current input was sufficiently predicted, OR
if we started over on startCells.
False indicates that the current input was NOT predicted,
and we are now bursting on most columns.
*/

func (tp *TemporalPooler) inferPhase1(activeColumns []int, useStartCells bool) bool {
	// Start with empty active state
	tp.DynamicState.infActiveState.Clear()

	// Phase 1 - turn on predicted cells in each column receiving bottom-up
	// If we are following a reset, activate only the start cell in each
	// column that has bottom-up
	numPredictedColumns := 0
	if useStartCells {
		for _, val := range activeColumns {
			tp.DynamicState.infActiveState.Set(val, 0, true)
		}
	} else {
		// else, turn on any predicted cells in each column. If there are none, then
		// turn on all cells (burst the column)
		for _, val := range activeColumns {
			predictingCells := tp.DynamicState.infPredictedStateLast.GetRowIndices(val)
			numPredictingCells := len(predictingCells)

			if numPredictingCells > 0 {
				//may have to set instead of replace
				tp.DynamicState.infActiveState.ReplaceRowByIndices(val, predictingCells)
				numPredictedColumns++
			} else {
				tp.DynamicState.infActiveState.FillRow(val, true) // whole column bursts
			}
		}
	}

	// Did we predict this input well enough?
	return useStartCells || numPredictedColumns >= int(0.50*float64(len(activeColumns)))

}
