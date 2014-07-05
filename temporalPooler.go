package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/zacg/floats"
	"github.com/zacg/go.matrix"
	//"math"
	"math/rand"
	//"sort"
)

type TpOutputType int

const (
	Normal                 TpOutputType = 0
	ActiveState            TpOutputType = 1
	ActiveState1CellPerCol TpOutputType = 2
)

type ProcessAction int

const (
	Update ProcessAction = 0
	Keep   ProcessAction = 1
	Remove ProcessAction = 2
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
	GlobalDecay            float64
	ActivationThreshold    int
	DoPooling              bool
	SegUpdateValidDuration int
	BurnIn                 int
	CollectStats           bool
	//Seed                   int
	Verbosity int
	//checkSynapseConsistency=False, # for cpp only -- ignored
	TrivialPredictionMethods []PredictorMethod
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
	if ds.cellConfidenceCandidate != nil {
		result.cellConfidenceCandidate = ds.cellConfidenceCandidate.Copy()
	}
	if ds.cellConfidenceLast != nil {
		result.cellConfidenceLast = ds.cellConfidenceLast.Copy()
	}

	result.colConfidence = make([]float64, len(ds.colConfidence))
	copy(result.colConfidence, ds.colConfidence)
	result.colConfidenceCandidate = make([]float64, len(ds.colConfidenceCandidate))
	copy(result.colConfidenceCandidate, ds.colConfidenceCandidate)
	result.colConfidenceLast = make([]float64, len(ds.colConfidenceLast))
	copy(result.colConfidenceLast, ds.colConfidenceLast)

	return result
}

type TemporalPooler struct {
	params          TemporalPoolerParams
	numberOfCells   int
	activeColumns   []int
	cells           [][][]Segment
	lrnIterationIdx int
	iterationIdx    int
	segId           int
	CurrentOutput   *SparseBinaryMatrix
	pamCounter      int
	avgInputDensity float64
	// Keeps track of the moving average of all learned sequence length.
	avgLearnedSeqLength float64
	resetCalled         bool
	// Keeps track of the length of the sequence currently being learned.
	learnedSeqLength     int
	trivialPredictor     *TrivialPredictor
	collectSequenceStats bool
	internalStats        *TpStats

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

//Initializes tp params with default values
func NewTemporalPoolerParams() *TemporalPoolerParams {
	tps := new(TemporalPoolerParams)

	tps.NumberOfCols = 500
	tps.CellsPerColumn = 10
	tps.InitialPerm = 0.11
	tps.ConnectedPerm = 0.50
	tps.MinThreshold = 8
	tps.NewSynapseCount = 15
	tps.PermanenceInc = 0.10
	tps.PermanenceDec = 0.10
	tps.PermanenceMax = 1.0
	tps.GlobalDecay = 0.10
	tps.ActivationThreshold = 12
	tps.DoPooling = false
	tps.SegUpdateValidDuration = 5
	tps.BurnIn = 2
	tps.CollectStats = false
	tps.Verbosity = 3
	//tps.TrivialPredictionMethods =
	tps.PamLength = 1
	tps.MaxInfBacktrack = 10
	tps.MaxLrnBacktrack = 5
	tps.MaxAge = 100000
	tps.MaxSeqLength = 32
	tps.MaxSegmentsPerCell = -1
	tps.MaxSynapsesPerSegment = -1
	tps.outputType = Normal

	return tps
}

//Initializes a new temporal pooler
func NewTemporalPooler(tParams TemporalPoolerParams) *TemporalPooler {
	tp := new(TemporalPooler)
	tp.params = tParams

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
	tp.cells = make([][][]Segment, tParams.NumberOfCols)
	for c := 0; c < tParams.NumberOfCols; c++ {
		tp.cells[c] = make([][]Segment, tParams.CellsPerColumn)

		// for i := 0; i < tParams.CellsPerColumn; i++ {
		// }
	}

	tp.lrnIterationIdx = 0
	tp.iterationIdx = 0
	tp.segId = 0
	tp.avgInputDensity = -1

	// pamCounter gets reset to pamLength whenever we detect that the learning
	// state is making good predictions (at least half the columns predicted).
	// Whenever we do not make a good prediction, we decrement pamCounter.
	// When pamCounter reaches 0, we start the learn state over again at start
	// cells.
	tp.pamCounter = tParams.PamLength

	// Trivial prediction algorithms

	if len(tParams.TrivialPredictionMethods) > 0 {
		tp.trivialPredictor = MakeTrivialPredictor(tParams.NumberOfCols, tParams.TrivialPredictionMethods)
	} else {
		tp.trivialPredictor = nil
	}

	// If True, the TP will compute a signature for each sequence
	tp.collectSequenceStats = false

	// This gets set when we receive a reset and cleared on the first compute
	// following a reset.
	tp.resetCalled = false

	tp.DynamicState = new(DynamicState)
	tp.DynamicState.infActiveState = NewSparseBinaryMatrix(tParams.NumberOfCols, tParams.CellsPerColumn)
	tp.DynamicState.infPredictedState = NewSparseBinaryMatrix(tParams.NumberOfCols, tParams.CellsPerColumn)
	tp.DynamicState.lrnActiveState = NewSparseBinaryMatrix(tParams.NumberOfCols, tParams.CellsPerColumn)
	tp.DynamicState.lrnPredictedState = NewSparseBinaryMatrix(tParams.NumberOfCols, tParams.CellsPerColumn)
	tp.DynamicState.cellConfidence = matrix.Zeros(tParams.NumberOfCols, tParams.CellsPerColumn)
	tp.DynamicState.colConfidence = make([]float64, tParams.NumberOfCols)

	tp.internalStats = new(TpStats)

	return tp
}

//Returns new unique segment id
func (tp *TemporalPooler) GetSegId() int {
	result := tp.segId
	tp.segId++
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
func (tp *TemporalPooler) columnConfidences() []float64 {
	//ignore cellconfidence param for now
	return tp.DynamicState.colConfidence
}

/*
 Top-down compute - generate expected input given output of the TP
	param topDownIn top down input from the level above us
	returns best estimate of the TP input that would have generated bottomUpOut.
*/
func (tp *TemporalPooler) topDownCompute() []float64 {
	/*
			 For now, we will assume there is no one above us and that bottomUpOut is
		     simply the output that corresponds to our currently stored column
		     confidences.

		     Simply return the column confidences
	*/

	return tp.columnConfidences()
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
returns all the future predictions - floating point matrix and
shape (nSteps, numberOfCols).
The ith row gives the tp prediction for each column at
a future timestep (t+i+1).
*/
func (tp *TemporalPooler) Predict(nSteps int) *matrix.DenseMatrix {
	// Save the TP dynamic state, we will use to revert back in the end
	pristineTPDynamicState := tp.DynamicState.Copy()

	if nSteps <= 0 {
		panic("nSteps must be greater than zero")
	}

	// multiStepColumnPredictions holds all the future prediction.
	elements := make([]float64, nSteps*tp.params.NumberOfCols)
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
func (tp *TemporalPooler) getSegmentActivityLevel(seg Segment, activeState *SparseBinaryMatrix,
	connectedSynapsesOnly bool) int {
	activity := 0
	//fmt.Println("syn count", len(seg.syns))
	if connectedSynapsesOnly {
		for _, val := range seg.syns {
			//fmt.Printf("***syn %v %v \n", val.Permanence, tp.params.ConnectedPerm)
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
				//fmt.Println("active syns", numActiveSyns)
				if numActiveSyns < tp.params.ActivationThreshold {
					continue
				}

				//Incorporate the confidence into the owner cell and column
				if tp.params.Verbosity >= 6 {
					fmt.Printf("incorporating DC from cell[%v,%v]: \n", c, i)
				}

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

/*
 This "backtracks" our inference state, trying to see if we can lock onto
the current set of inputs by assuming the sequence started up to N steps
ago on start cells.

param activeColumns The list of active column indices

This will adjustinfActiveState['t'] if it does manage to lock on to a
sequence that started earlier. It will also compute infPredictedState['t']
based on the possibly updatedinfActiveState['t'], so there is no need to
call inferPhase2() after calling inferBacktrack().

This looks at:
- infActiveState['t']

This updates/modifies:
-infActiveState['t']
-infPredictedState['t']
-colConfidence['t']
-cellConfidence['t']

How it works:
-------------------------------------------------------------------
This method gets called from updateInferenceState when we detect either of
the following two conditions:
- The current bottom-up input had too many un-expected columns
- We fail to generate a sufficient number of predicted columns for the
next time step.

Either of these two conditions indicate that we have fallen out of a
learned sequence.

Rather than simply "giving up" and bursting on the unexpected input
columns, a better approach is to see if perhaps we are in a sequence that
started a few steps ago. The real world analogy is that you are driving
along and suddenly hit a dead-end, you will typically go back a few turns
ago and pick up again from a familiar intersection.

This back-tracking goes hand in hand with our learning methodology, which
always tries to learn again from start cells after it loses context. This
results in a network that has learned multiple, overlapping paths through
the input data, each starting at different points. The lower the global
decay and the more repeatability in the data, the longer each of these
paths will end up being.

The goal of this function is to find out which starting point in the past
leads to the current input with the most context as possible. This gives us
the best chance of predicting accurately going forward. Consider the
following example, where you have learned the following sub-sequences which
have the given frequencies:

? - Q - C - D - E 10X seq 0
? - B - C - D - F 1X seq 1
? - B - C - H - I 2X seq 2
? - B - C - D - F 3X seq 3
? - Z - A - B - C - D - J 2X seq 4
? - Z - A - B - C - H - I 1X seq 5
? - Y - A - B - C - D - F 3X seq 6

----------------------------------------
W - X - Z - A - B - C - D <= input history
^
current time step

Suppose, in the current time step, the input pattern is D and you have not
predicted D, so you need to backtrack. Suppose we can backtrack up to 6
steps in the past, which path should we choose? From the table above, we can
see that the correct answer is to assume we are in seq 1. How do we
implement the backtrack to give us this right answer? The current
implementation takes the following approach:

- Start from the farthest point in the past.
- For each starting point S, calculate the confidence of the current
input, conf(startingPoint=S), assuming we followed that sequence.
Note that we must have learned at least one sequence that starts at
point S.
- If conf(startingPoint=S) is significantly different from
conf(startingPoint=S-1), then choose S-1 as the starting point.

The assumption here is that starting point S-1 is the starting point of
a learned sub-sequence that includes the current input in it's path and
that started the longest ago. It thus has the most context and will be
the best predictor going forward.

From the statistics in the above table, we can compute what the confidences
will be for each possible starting point:

startingPoint confidence of D
-----------------------------------------
B (t-2) 4/6 = 0.667 (seq 1,3)/(seq 1,2,3)
Z (t-4) 2/3 = 0.667 (seq 4)/(seq 4,5)

First of all, we do not compute any confidences at starting points t-1, t-3,
t-5, t-6 because there are no learned sequences that start at those points.

Notice here that Z is the starting point of the longest sub-sequence leading
up to the current input. Event though starting at t-2 and starting at t-4
give the same confidence value, we choose the sequence starting at t-4
because it gives the most context, and it mirrors the way that learning
extends sequences.
*/
func (tp *TemporalPooler) inferBacktrack(activeColumns []int) {
	// How much input history have we accumulated?
	// The current input is always at the end of self._prevInfPatterns (at
	// index -1), but it is also evaluated as a potential starting point by
	// turning on it's start cells and seeing if it generates sufficient
	// predictions going forward.
	numPrevPatterns := len(tp.prevInfPatterns)
	if numPrevPatterns <= 0 {
		return
	}

	// This is an easy to use label for the current time step
	currentTimeStepsOffset := numPrevPatterns - 1

	// Save our current active state in case we fail to find a place to restart
	// todo: save infActiveState['t-1'], infPredictedState['t-1']?
	tp.DynamicState.infActiveStateBackup = tp.DynamicState.infActiveState.Copy()

	// Save our t-1 predicted state because we will write over it as as evaluate
	// each potential starting point.
	tp.DynamicState.infPredictedStateBackup = tp.DynamicState.infPredictedStateLast.Copy()

	// We will record which previous input patterns did not generate predictions
	// up to the current time step and remove all the ones at the head of the
	// input history queue so that we don't waste time evaluating them again at
	// a later time step.
	var badPatterns []int

	// Let's go back in time and replay the recent inputs from start cells and
	// see if we can lock onto this current set of inputs that way.

	// Start the farthest back and work our way forward. For each starting point,
	// See if firing on start cells at that point would predict the current
	// input as well as generate sufficient predictions for the next time step.

	// We want to pick the point closest to the current time step that gives us
	// the relevant confidence. Think of this example, where we are at D and need
	// to
	// A - B - C - D
	// decide if we should backtrack to C, B, or A. Suppose B-C-D is a high order
	// sequence and A is unrelated to it. If we backtrock to B would we get a
	// certain confidence of D, but if went went farther back, to A, the
	// confidence wouldn't change, since A has no impact on the B-C-D series.

	// So, our strategy will be to pick the "B" point, since choosing the A point
	// does not impact our confidences going forward at all.
	inSequence := false
	candConfidence := -1.0
	candStartOffset := 0

	//for startOffset in range(0, numPrevPatterns):
	for startOffset := 0; startOffset < numPrevPatterns; startOffset++ {
		// If we have a candidate already in the past, don't bother falling back
		// to start cells on the current input.
		if startOffset == currentTimeStepsOffset && candConfidence != -1 {
			break
		}

		if tp.params.Verbosity >= 3 {
			fmt.Printf("Trying to lock-on using startCell state from %v steps ago: %v \n",
				(numPrevPatterns - 1 - startOffset), tp.prevInfPatterns[startOffset])
		}

		// Play through starting from starting point 'startOffset'
		inSequence = false
		totalConfidence := 0.0
		//for offset in range(startOffset, numPrevPatterns):
		for offset := startOffset; offset < numPrevPatterns; offset++ {
			// If we are about to set the active columns for the current time step
			// based on what we predicted, capture and save the total confidence of
			// predicting the current input

			if offset == currentTimeStepsOffset {
				for _, val := range activeColumns {
					totalConfidence += tp.DynamicState.colConfidence[val]
				}
			}

			// Compute activeState[t] given bottom-up and predictedState  t-1
			tp.DynamicState.infPredictedStateLast = tp.DynamicState.infPredictedState.Copy()

			inSequence = tp.inferPhase1(tp.prevInfPatterns[offset], (offset == startOffset))
			if !inSequence {
				break
			}

			if tp.params.Verbosity >= 3 {
				fmt.Println("backtrack: computing predictions from", tp.prevInfPatterns[offset])
			}

			// Compute predictedState at t given activeState at t
			inSequence = tp.inferPhase2()
			if !inSequence {
				break
			}

		}

		// If starting from startOffset got lost along the way, mark it as an
		// invalid start point.
		if !inSequence {
			badPatterns = append(badPatterns, startOffset)
			continue
		}

		// If we got to here, startOffset is a candidate starting point.
		// Save this state as a candidate state. It will become the chosen state if
		// we detect a change in confidences starting at a later startOffset
		candConfidence = totalConfidence
		candStartOffset = startOffset

		if tp.params.Verbosity >= 3 &&
			startOffset != currentTimeStepsOffset {
			fmt.Printf("# Prediction confidence of current input after starting %v steps ago: %v \n", (numPrevPatterns - 1 - startOffset), totalConfidence)
		}

		if candStartOffset == currentTimeStepsOffset { // no more to try
			break
		}
		fmt.Println("Setting candidates")
		tp.DynamicState.infActiveStateCandidate = tp.DynamicState.infActiveState.Copy()
		tp.DynamicState.infPredictedStateCandidate = tp.DynamicState.infPredictedState.Copy()
		tp.DynamicState.cellConfidenceCandidate = tp.DynamicState.cellConfidence.Copy()

		tp.DynamicState.colConfidenceCandidate = make([]float64, len(tp.DynamicState.colConfidence))
		copy(tp.DynamicState.colConfidenceCandidate, tp.DynamicState.colConfidence)
		break

	}

	// If we failed to lock on at any starting point, fall back to the original
	// active state that we had on entry
	if candStartOffset == -1 {
		if tp.params.Verbosity >= 3 {
			fmt.Println("Failed to lock on. Falling back to bursting all unpredicted.")
		}
		tp.DynamicState.infActiveState = tp.DynamicState.infActiveStateBackup
		tp.inferPhase2()
	} else {

		if tp.params.Verbosity >= 3 {
			fmt.Printf("Locked on to current input by using start cells from %v steps ago: %v \n",
				(numPrevPatterns - 1 - candStartOffset), tp.prevInfPatterns[candStartOffset])
		}

		fmt.Printf("candstart=%v currentTimeStepoff=%v \n", candStartOffset, currentTimeStepsOffset)

		// Install the candidate state, if it wasn't the last one we evaluated.
		if candStartOffset != currentTimeStepsOffset {
			fmt.Println("Installing candidate state...")
			tp.DynamicState.infActiveState = tp.DynamicState.infActiveStateCandidate.Copy()
			tp.DynamicState.infPredictedState = tp.DynamicState.infPredictedStateCandidate.Copy()
			tp.DynamicState.cellConfidence = tp.DynamicState.cellConfidenceCandidate.Copy()
			copy(tp.DynamicState.colConfidence, tp.DynamicState.colConfidenceCandidate)
			//tp.DynamicState.colConfidence = tp.DynamicState.colConfidenceCandidate
		}

	}

	// Remove any useless patterns at the head of the previous input pattern
	// queue.
	for i := 0; i < numPrevPatterns; i++ {
		if ContainsInt(i, badPatterns) || (candStartOffset != -1 && i <= candStartOffset) {

			if tp.params.Verbosity >= 3 {
				fmt.Printf("Removing useless pattern 1 of %v from history: %v \n", len(tp.prevInfPatterns), tp.prevInfPatterns[0])
			}

			//pop prev pattern
			tp.prevInfPatterns = tp.prevInfPatterns[:len(tp.prevInfPatterns)-1]
		} else {
			break
		}
	}

	// Restore the original predicted state.
	tp.DynamicState.infPredictedState = tp.DynamicState.infPredictedStateBackup.Copy()
}

/*
 Update the inference state. Called from compute() on every iteration.
param activeColumns The list of active column indices.
*/
func (tp *TemporalPooler) updateInferenceState(activeColumns []int) {

	// Copy t to t-1
	tp.DynamicState.infActiveStateLast = tp.DynamicState.infActiveState.Copy()
	tp.DynamicState.infPredictedStateLast = tp.DynamicState.infPredictedState.Copy()
	tp.DynamicState.cellConfidenceLast = tp.DynamicState.cellConfidence.Copy()
	tp.DynamicState.colConfidenceLast = make([]float64, len(tp.DynamicState.colConfidence))
	copy(tp.DynamicState.colConfidenceLast, tp.DynamicState.colConfidence)

	// Each phase will zero/initilize the 't' states that it affects

	// Update our inference input history
	if tp.params.MaxInfBacktrack > 0 {
		if len(tp.prevInfPatterns) > tp.params.MaxInfBacktrack {
			//pop prev pattern
			tp.prevInfPatterns = tp.prevInfPatterns[:len(tp.prevInfPatterns)-1]
		}

		tp.prevInfPatterns = append(tp.prevInfPatterns, activeColumns)
	}

	// Compute the active state given the predictions from last time step and
	// the current bottom-up
	inSequence := tp.inferPhase1(activeColumns, tp.resetCalled)

	// If this input was considered unpredicted, let's go back in time and
	// replay the recent inputs from start cells and see if we can lock onto
	// this current set of inputs that way.
	if !inSequence {
		if tp.params.Verbosity >= 3 {
			fmt.Println("Too much unpredicted input, re-tracing back to try and lock on at an earlier timestep.")
		}

		// inferBacktrack() will call inferPhase2() for us.
		tp.inferBacktrack(activeColumns)
		return
	}

	// Compute the predicted cells and the cell and column confidences
	inSequence = tp.inferPhase2()

	if !inSequence {
		if tp.params.Verbosity >= 3 {
			fmt.Println("Not enough predictions going forward, re-tracing back to try and lock on at an earlier timestep.")
		}

		// inferBacktrack() will call inferPhase2() for us.
		tp.inferBacktrack(activeColumns)
	}

}

/*
Remove a segment update (called when seg update expires or is processed)
*/
func (tp *TemporalPooler) removeSegmentUpdate(updateState UpdateState) {
	// Key is stored in segUpdate itself...
	key := TupleInt{updateState.Update.columnIdx, updateState.Update.cellIdx}
	delete(tp.segmentUpdates, key)
}

/*
 Removes any update that would be for the given col, cellIdx, segIdx.
NOTE: logically, we need to do this when we delete segments, so that if
an update refers to a segment that was just deleted, we also remove
that update from the update list. However, I haven't seen it trigger
in any of the unit tests yet, so it might mean that it's not needed
and that situation doesn't occur, by construction.
*/
func (tp *TemporalPooler) cleanUpdatesList(col, cellIdx int, seg Segment) {
	for idx, val := range tp.segmentUpdates {
		if idx.A == col && idx.B == cellIdx {
			for _, update := range val {
				if update.Update.segment.Equals(&seg) {
					tp.removeSegmentUpdate(update)
				}
			}
		}
	}

}

/*
 This method goes through a list of segments for a given cell and
deletes all synapses whose permanence is less than minPermanence and deletes
any segments that have less than minNumSyns synapses remaining.

param colIdx Column index
param cellIdx Cell index within the column
param segList List of segment references
param minPermanence Any syn whose permamence is 0 or < minPermanence will
be deleted.
param minNumSyns Any segment with less than minNumSyns synapses remaining
in it will be deleted.

returns tuple (numSegsRemoved, numSynsRemoved)
*/
func (tp *TemporalPooler) trimSegmentsInCell(colIdx, cellIdx int, segList []Segment,
	minPermanence float64, minNumSyns int) (int, int) {

	// Fill in defaults
	//minPermanence = tp.connectedPerm
	//minNumSyns = tp.activationThreshold

	// Loop through all segments
	nSegsRemoved, nSynsRemoved := 0, 0
	var segsToDel []Segment // collect and remove segments outside the loop

	for _, segment := range segList {
		// List if synapses to delete
		var synsToDel []Synapse

		for _, syn := range segment.syns {
			if syn.Permanence < minPermanence {
				synsToDel = append(synsToDel, syn)
			}
		}

		nSynsRemoved := 0
		if len(synsToDel) == len(segment.syns) {
			segsToDel = append(segsToDel, segment) // will remove the whole segment
		} else {
			if len(synsToDel) > 0 {
				var temp []Synapse
				for _, osyn := range segment.syns {
					found := false
					for _, syn := range synsToDel {
						if syn == osyn {
							found = true
							break
						}
						nSynsRemoved++
					}
					if !found {
						temp = append(temp, osyn)
					}
				}
				segment.syns = temp
				nSynsRemoved += len(synsToDel)

				//for syn in synsToDel: // remove some synapses on segment
				//	segment.syns.remove(syn)

			}

			if len(segment.syns) < minNumSyns {
				segsToDel = append(segsToDel, segment)
			}
		}
	}

	// Remove segments that don't have enough synapses and also take them
	// out of the segment update list, if they are in there
	nSegsRemoved += len(segsToDel)

	// remove some segments of this cell
	for _, seg := range segsToDel {
		tp.cleanUpdatesList(colIdx, cellIdx, seg)
		for idx, val := range tp.cells[colIdx][cellIdx] {
			if val.Equals(&seg) {
				copy(tp.cells[colIdx][cellIdx][idx:], tp.cells[colIdx][cellIdx][idx+1:])
				tp.cells[colIdx][cellIdx][len(tp.cells[colIdx][cellIdx])-1] = Segment{}
				tp.cells[colIdx][cellIdx] = tp.cells[colIdx][cellIdx][:len(tp.cells[colIdx][cellIdx])-1]
				break
			}
		}
		nSynsRemoved += len(seg.syns)
	}

	return nSegsRemoved, nSynsRemoved
}

/*
 Go through the list of accumulated segment updates and process them
as follows:

if the segment update is too old, remove the update
else if the cell received bottom-up, update its permanences
else if it's still being predicted, leave it in the queue
else remove it.
*/
func (tp *TemporalPooler) processSegmentUpdates(activeColumns []int) {
	// The segmentUpdates dict has keys which are the column,cellIdx of the
	// owner cell. The values are lists of segment updates for that cell
	var removeKeys []TupleInt
	var trimSegments []UpdateState

	fmt.Println("seg updates:", len(tp.segmentUpdates))

	for key, updateList := range tp.segmentUpdates {
		// Get the column number and cell index of the owner cell
		var action ProcessAction

		// If the cell received bottom-up, update its segments
		if ContainsInt(key.A, activeColumns) {
			action = Update
		} else {
			// If not, either keep it around if it's still predicted, or remove it
			// If it is still predicted, and we are pooling, keep it around
			if tp.params.DoPooling && tp.DynamicState.lrnPredictedState.Get(key.A, key.B) {
				action = Keep
			} else {
				action = Remove
			}
		}

		fmt.Println("action:", action)

		// Process each segment for this cell. Each segment entry contains
		// [creationDate, SegmentState]
		var updateListKeep []UpdateState
		if action != Remove {
			for _, updateState := range updateList {

				if tp.params.Verbosity >= 4 {
					fmt.Println("nLrnIterations=", tp.lrnIterationIdx)
					fmt.Println(updateState)
				}

				// If this segment has expired. Ignore this update (and hence remove it
				// from list)
				if tp.lrnIterationIdx-updateState.CreationDate > tp.params.SegUpdateValidDuration {
					continue
				}

				if action == Update {
					fmt.Println("updating segment...")
					trimSegment := updateState.Update.adaptSegments(tp)
					if trimSegment {
						trimSegments = append(trimSegments, updateState)
					}
				} else {
					fmt.Println("keeping segment...")
					// Keep segments that haven't expired yet (the cell is still being
					// predicted)
					updateListKeep = append(updateListKeep, updateState)

				}

			}
		}

		tp.segmentUpdates[key] = updateListKeep
		if len(updateListKeep) == 0 {
			removeKeys = append(removeKeys, key)
		}

	} //end segment update loop

	// Clean out empty segment updates
	for _, k := range removeKeys {
		delete(tp.segmentUpdates, k)
	}

	// Trim segments that had synapses go to 0
	for _, val := range trimSegments {
		ud := val.Update
		tp.trimSegmentsInCell(ud.columnIdx, ud.cellIdx, []Segment{*ud.segment}, 0.00001, 0)
	}

}

/*
 Find weakly activated cell in column with at least minThreshold active
synapses.

param c which column to look at
param activeState the active cells
param minThreshold minimum number of synapses required

returns tuple (cellIdx, segment, numActiveSynapses)
*/
func (tp *TemporalPooler) getBestMatchingCell(c int, activeState *SparseBinaryMatrix, minThreshold int) (int, *Segment, int) {
	// Collect all cells in column c that have at least minThreshold in the most
	// activated segment
	bestActivityInCol := minThreshold
	bestSegIdxInCol := -1
	bestCellInCol := -1

	for i := 0; i < tp.params.CellsPerColumn; i++ {
		maxSegActivity := 0
		maxSegIdx := 0

		for idx, s := range tp.cells[c][i] {
			activity := tp.getSegmentActivityLevel(s, activeState, false)

			if activity > maxSegActivity {
				maxSegActivity = activity
				maxSegIdx = idx
			}

		}

		if maxSegActivity >= bestActivityInCol {
			bestActivityInCol = maxSegActivity
			bestSegIdxInCol = maxSegIdx
			bestCellInCol = i
		}

	}

	if bestCellInCol == -1 {
		return -1, nil, -1
	} else {
		return bestCellInCol, &tp.cells[c][bestCellInCol][bestSegIdxInCol], bestActivityInCol
	}

}

/*
 Choose n random cells to learn from.

 This function is called several times while learning with timeStep = t-1, so
 we cache the set of candidates for that case. It's also called once with
 timeStep = t, and we cache that set of candidates.
*/
func (tp *TemporalPooler) chooseCellsToLearnFrom(s *Segment, n int,
	activeState *SparseBinaryMatrix) []SparseEntry {
	if n <= 0 {
		return nil
	}

	//fmt.Printf("entering chooseCells n:%v ae:%v snil:%v \n", n, len(activeState.Entries), s == nil)

	// Candidates can be empty at this point, in which case we return
	// an empty segment list. adaptSegments will do nothing when getting
	// that list.
	if len(activeState.Entries) == 0 {
		return nil
	}

	var candidates []SparseEntry

	if s != nil {
		// We exclude any synapse that is already in this segment.
		for _, cand := range activeState.Entries {
			found := false
			for _, syn := range s.syns {
				if syn.SrcCellCol == cand.Col &&
					syn.SrcCellIdx == cand.Row {
					found = true
					break
				}
			}
			if !found {
				candidates = append(candidates, cand)
			}
		}
	} else {
		candidates = make([]SparseEntry, len(activeState.Entries))
		copy(candidates, activeState.Entries)
	}

	// If we have no more candidates than requested, return all of them,
	// no shuffle necessary.
	if len(candidates) <= n {
		return candidates
	}

	//if only one is required pick a random candidate
	if n == 1 {
		idx := rand.Intn(len(candidates))
		return []SparseEntry{candidates[idx]} // col and cell idx in col
	}

	// If we need more than one candidate pick a random selection
	size := mathutil.Min(n, len(candidates))

	//Shuffle candidates
	for i := range candidates {
		j := rand.Intn(i + 1)
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}

	return candidates[:size]
}

/*
 Return the index of a cell in this column which is a good candidate
for adding a new segment.

When we have fixed size resources in effect, we insure that we pick a
cell which does not already have the max number of allowed segments. If
none exists, we choose the least used segment in the column to re-allocate.

param colIdx which column to look at
returns cell index
*/
func (tp *TemporalPooler) getCellForNewSegment(colIdx int) int {
	// Not fixed size CLA, just choose a cell randomly
	if tp.params.MaxSegmentsPerCell < 0 {
		i := 0
		if tp.params.CellsPerColumn > 1 {
			// Don't ever choose the start cell (cell # 0) in each column
			i = rand.Intn(tp.params.CellsPerColumn-1) + 1
		}
		return i
	}

	// Fixed size CLA, choose from among the cells that are below the maximum
	// number of segments.
	// NOTE: It is important NOT to always pick the cell with the fewest number
	// of segments. The reason is that if we always do that, we are more likely
	// to run into situations where we choose the same set of cell indices to
	// represent an 'A' in both context 1 and context 2. This is because the
	// cell indices we choose in each column of a pattern will advance in
	// lockstep (i.e. we pick cell indices of 1, then cell indices of 2, etc.).
	var candidateCellIdxs []int

	minIdx := 0
	maxIdx := 0
	if tp.params.CellsPerColumn != 1 {
		minIdx = 1 // Don't include startCell in the mix
		maxIdx = tp.params.CellsPerColumn - 1
	}

	for i := minIdx; i <= maxIdx; i++ {
		numSegs := len(tp.cells[colIdx][i])
		if numSegs < tp.params.MaxSegmentsPerCell {
			candidateCellIdxs = append(candidateCellIdxs, i)
		}
	}

	// If we found one, return with it. Note we need to use _random to maintain
	// correspondence with CPP code.
	if len(candidateCellIdxs) > 0 {
		cellIdx := rand.Intn(len(candidateCellIdxs))
		if tp.params.Verbosity >= 5 {
			fmt.Printf("Cell [%v,%v] chosen for new segment, # of segs is %v \n",
				colIdx, candidateCellIdxs[cellIdx], len(tp.cells[colIdx][cellIdx]))
		}
		return candidateCellIdxs[cellIdx]
	}

	// All cells in the column are full, find a segment to free up
	var candidateSegment Segment
	candidateSegmentDC := 1.0
	candidateCellIdx := -1
	candidateSegIdx := -1
	// For each cell in this column
	for i := minIdx; i <= maxIdx; i++ {
		for idx, s := range tp.cells[colIdx][i] {
			dc := s.dutyCycle(false, false)
			if dc < candidateSegmentDC {
				candidateCellIdx = i
				candidateSegmentDC = dc
				candidateSegment = s
				candidateSegIdx = idx
			}
		}
	}

	// Free up the least used segment
	if tp.params.Verbosity >= 5 {
		fmt.Printf("Deleting segment #%v for cell[%v,%v] to make room for new %v segment \n",
			candidateSegment.segId, colIdx, candidateCellIdx)
		//fmt.Print(candidateSegment.ToString())
	}

	tp.cleanUpdatesList(colIdx, candidateCellIdx, candidateSegment)

	//delete segment from cells
	copy(tp.cells[colIdx][candidateCellIdx][candidateSegIdx:], tp.cells[colIdx][candidateCellIdx][candidateSegIdx+1:])
	tp.cells[colIdx][candidateCellIdx][len(tp.cells[colIdx][candidateCellIdx])-1] = Segment{}
	tp.cells[colIdx][candidateCellIdx] = tp.cells[colIdx][candidateCellIdx][:len(tp.cells[colIdx][candidateCellIdx])-1]

	return candidateCellIdx
}

/*
 For the given cell, find the segment with the largest number of active
synapses. This routine is aggressive in finding the best match. The
permanence value of synapses is allowed to be below connectedPerm. The number
of active synapses is allowed to be below activationThreshold, but must be
above minThreshold. The routine returns the segment
*/
func (tp *TemporalPooler) getBestMatchingSegment(c int, i int, activeState *SparseBinaryMatrix) *Segment {
	maxActivity := tp.params.MinThreshold
	which := -1

	for idx, s := range tp.cells[c][i] {
		activity := tp.getSegmentActivityLevel(s, activeState, false)
		if activity >= maxActivity {
			maxActivity = activity
			which = idx
		}
	}

	if which == -1 {
		return nil
	} else {
		return &tp.cells[c][i][which]
	}

}

/*
 Compute the learning active state given the predicted state and
the bottom-up input.

param activeColumns list of active bottom-ups
param readOnly True if being called from backtracking logic.
This tells us not to increment any segment
duty cycles or queue up any updates.
returns True if the current input was sufficiently predicted, OR
if we started over on startCells. False indicates that the current
input was NOT predicted, well enough to consider it as "inSequence"

This looks at:
-lrnActiveState['t-1']
-lrnPredictedState['t-1']

This modifies:
-lrnActiveState['t']
-lrnActiveState['t-1']

*/
func (tp *TemporalPooler) learnPhase1(activeColumns []int, readOnly bool) bool {
	// Save previous active state and start out on a clean slate
	tp.DynamicState.lrnActiveState.Clear()

	// For each column, turn on the predicted cell. There will always be at most
	// one predicted cell per column
	numUnpredictedColumns := 0

	for c := range activeColumns {
		predictingCells := tp.DynamicState.lrnPredictedState.GetRowIndices(c)
		numPredictedCells := len(predictingCells)
		if numPredictedCells > 1 {
			panic("number of predicted cells too high")
		}

		// If we have a predicted cell, turn it on. The segment's posActivation
		// count will have already been incremented by processSegmentUpdates
		if numPredictedCells == 1 {
			i := predictingCells[0]
			tp.DynamicState.lrnActiveState.Set(c, i, true)
			continue
		}

		numUnpredictedColumns += 1
		if readOnly {
			continue
		}

		// If no predicted cell, pick the closest matching one to reinforce, or
		// if none exists, create a new segment on a cell in that column
		i, s, _ := tp.getBestMatchingCell(c, tp.DynamicState.lrnActiveStateLast, tp.params.MinThreshold)

		if s != nil && s.isSequenceSeg {
			if tp.params.Verbosity >= 4 {
				fmt.Printf("Learn branch 0, found segment match. Learning on col= %v \n", c)
			}

			tp.DynamicState.lrnActiveState.Set(c, i, true)
			segUpdate := tp.getSegmentActiveSynapses(c, i, s, tp.DynamicState.lrnActiveStateLast, true)
			s.totalActivations++
			// This will update the permanences, posActivationsCount, and the
			// lastActiveIteration (age).
			trimSegment := segUpdate.adaptSegments(tp)

			if trimSegment {
				tp.trimSegmentsInCell(c, i, []Segment{*s}, 0.00001, 0)
			}

		} else {
			// If no close match exists, create a new one
			// Choose a cell in this column to add a new segment to
			i = tp.getCellForNewSegment(c)
			if tp.params.Verbosity >= 4 {
				fmt.Printf("Learn branch 1, no match. Learning on col=%v , newCellIdxInCol=%v \n", c, i)
			}

			tp.DynamicState.lrnActiveState.Set(c, i, true)
			segUpdate := tp.getSegmentActiveSynapses(c, i, nil, tp.DynamicState.lrnActiveStateLast, true)
			segUpdate.sequenceSegment = true
			segUpdate.adaptSegments(tp) // No need to check whether perm reached 0
		}

	}

	// Determine if we are out of sequence or not and reset our PAM counter
	// if we are in sequence
	numBottomUpColumns := len(activeColumns)

	//true if in sequence, false if out of sequence
	return numUnpredictedColumns < numBottomUpColumns/2

}

/*
 Compute the predicted segments given the current set of active cells.

param readOnly True if being called from backtracking logic.
This tells us not to increment any segment
duty cycles or queue up any updates.

This computes the lrnPredictedState['t'] and queues up any segments that
became active (and the list of active synapses for each segment) into
the segmentUpdates queue

This looks at:
-lrnActiveState['t']

This modifies:
-lrnPredictedState['t']
-segmentUpdates
*/
func (tp *TemporalPooler) learnPhase2(readOnly bool) {
	// Clear out predicted state to start with
	tp.DynamicState.lrnPredictedState.Clear()

	// Compute new predicted state. When computing predictions for
	// phase 2, we predict at most one cell per column (the one with the best
	// matching segment).

	for c := 0; c < tp.params.NumberOfCols; c++ {
		// Is there a cell predicted to turn on in this column?
		i, s, numActive := tp.getBestMatchingCell(c, tp.DynamicState.lrnActiveState, tp.params.ActivationThreshold)
		if i == -1 {
			continue
		}

		// Turn on the predicted state for the best matching cell and queue
		// the pertinent segment up for an update, which will get processed if
		// the cell receives bottom up in the future.
		tp.DynamicState.lrnPredictedState.Set(c, i, true)
		if readOnly {
			continue
		}

		//Queue up this segment for updating
		newSyns := numActive < tp.params.NewSynapseCount
		segUpdate := tp.getSegmentActiveSynapses(c, i, s, tp.DynamicState.lrnActiveState, newSyns)

		s.totalActivations++ // increment totalActivations
		tp.addToSegmentUpdates(c, i, segUpdate)

		if tp.params.DoPooling {
			// creates a new pooling segment if no best matching segment found
			// sum(all synapses) >= minThreshold, "weak" activation
			predSegment := tp.getBestMatchingSegment(c, i, tp.DynamicState.lrnActiveStateLast)
			segUpdate := tp.getSegmentActiveSynapses(c, i, predSegment, tp.DynamicState.lrnActiveStateLast, true)
			tp.addToSegmentUpdates(c, i, segUpdate)
		}

	}

}

/*
 A utility method called from learnBacktrack. This will backtrack
starting from the given startOffset in our prevLrnPatterns queue.

It returns True if the backtrack was successful and we managed to get
predictions all the way up to the current time step.

If readOnly, then no segments are updated or modified, otherwise, all
segment updates that belong to the given path are applied.

This updates/modifies:
- lrnActiveState['t']

This trashes:
- lrnPredictedState['t']
- lrnPredictedState['t-1']
- lrnActiveState['t-1']

param startOffset Start offset within the prevLrnPatterns input history
returns True if we managed to lock on to a sequence that started
earlier.
If False, we lost predictions somewhere along the way
leading up to the current time.
*/
func (tp *TemporalPooler) learnBacktrackFrom(startOffset int, readOnly bool) bool {
	// How much input history have we accumulated?
	// The current input is always at the end of self._prevInfPatterns (at
	// index -1), but it is also evaluated as a potential starting point by
	// turning on it's start cells and seeing if it generates sufficient
	// predictions going forward.
	numPrevPatterns := len(tp.prevLrnPatterns)

	// This is an easy to use label for the current time step
	currentTimeStepsOffset := numPrevPatterns - 1

	// Clear out any old segment updates. learnPhase2() adds to the segment
	// updates if we're not readOnly
	if !readOnly {
		tp.segmentUpdates = nil
	}

	// Status message

	if tp.params.Verbosity >= 3 {
		if readOnly {
			fmt.Printf("Trying to lock-on using startCell state from %v steps ago: %v \n",
				(numPrevPatterns - 1 - startOffset), tp.prevLrnPatterns[startOffset])
		} else {
			fmt.Printf("Locking on using startCell state from %v steps ago: %v \n",
				(numPrevPatterns - 1 - startOffset), tp.prevLrnPatterns[startOffset])
		}
	}

	// Play through up to the current time step
	inSequence := true
	for offset := startOffset; offset < numPrevPatterns; offset++ {
		// Copy predicted and active states into t-1
		tp.DynamicState.lrnPredictedStateLast = tp.DynamicState.lrnPredictedState.Copy()
		tp.DynamicState.lrnActiveStateLast = tp.DynamicState.lrnActiveState.Copy()

		// Get the input pattern
		inputColumns := tp.prevLrnPatterns[offset]

		// Apply segment updates from the last set of predictions
		if !readOnly {
			tp.processSegmentUpdates(inputColumns)
		}

		// Phase 1:
		// Compute activeState[t] given bottom-up and predictedState[t-1]
		if offset == startOffset {
			tp.DynamicState.lrnActiveState.Clear()
			for c := range inputColumns {
				tp.DynamicState.lrnActiveState.Set(c, 0, true)
			}
			inSequence = true
		} else {
			// Uses lrnActiveState['t-1'] and lrnPredictedState['t-1']
			// computes lrnActiveState['t']
			inSequence = tp.learnPhase1(inputColumns, readOnly)
		}

		// Break out immediately if we fell out of sequence or reached the current
		// time step
		if !inSequence || offset == currentTimeStepsOffset {
			break
		}

		if tp.params.Verbosity >= 3 {
			fmt.Println("backtrack: computing predictions from:", inputColumns)
		}

		// Phase 2:
		// Computes predictedState['t'] given activeState['t'] and also queues
		// up active segments into self.segmentUpdates, unless this is readOnly
		tp.learnPhase2(readOnly)

	}

	// Return whether or not this starting point was valid
	return inSequence
}

/*
	 This "backtracks" our learning state, trying to see if we can lock onto
the current set of inputs by assuming the sequence started up to N steps
ago on start cells.

This will adjustlrnActiveState['t'] if it does manage to lock on to a
sequence that started earlier.

returns >0 if we managed to lock on to a sequence that started
earlier. The value returned is how many steps in the
past we locked on.
If 0 is returned, the caller needs to change active
state to start on start cells.

How it works:
-------------------------------------------------------------------
This method gets called from updateLearningState when we detect either of
the following two conditions:
- Our PAM counter (pamCounter) expired
- We reached the max allowed learned sequence length

Either of these two conditions indicate that we want to start over on start
cells.

Rather than start over on start cells on the current input, we can
accelerate learning by backtracking a few steps ago and seeing if perhaps
a sequence we already at least partially know already started.

This updates/modifies:
-lrnActiveState['t']

This trashes:
-lrnActiveState['t-1']
-lrnPredictedState['t']
-lrnPredictedState['t-1']


*/
func (tp *TemporalPooler) learnBacktrack() int {

	// How much input history have we accumulated?
	// The current input is always at the end of prevInfPatterns (at
	// index -1), and is not a valid startingOffset to evaluate.
	numPrevPatterns := len(tp.prevLrnPatterns) - 1
	if numPrevPatterns <= 0 {
		if tp.params.Verbosity >= 3 {
			fmt.Println("lrnBacktrack: No available history to backtrack from")
		}
		return -1
	}

	// We will record which previous input patterns did not generate predictions
	// up to the current time step and remove all the ones at the head of the
	// input history queue so that we don't waste time evaluating them again at
	// a later time step.
	var badPatterns []int

	// Let's go back in time and replay the recent inputs from start cells and
	// see if we can lock onto this current set of inputs that way.
	//
	// Start the farthest back and work our way forward. For each starting point,
	// See if firing on start cells at that point would predict the current
	// input.
	//
	// We want to pick the point farthest in the past that has continuity
	// up to the current time step
	inSequence := false
	startOffset := 0
	for startOffset < numPrevPatterns {
		// Can we backtrack from startOffset?
		inSequence := tp.learnBacktrackFrom(startOffset, true)

		// Done playing through the sequence from starting point startOffset
		// Break out as soon as we find a good path
		if inSequence {
			break
		}

		// Take this bad starting point out of our input history so we don't
		// try it again later.
		badPatterns = append(badPatterns, startOffset)
		startOffset++
	}

	// If we failed to lock on at any starting point, return failure. The caller
	// will start over again on start cells
	if !inSequence {
		if tp.params.Verbosity >= 3 {
			fmt.Println("Failed to lock on. Falling back to start cells on current time step.")
		}

		// Nothing in our input history was a valid starting point, so get rid
		// of it so we don't try any of them again at a later iteration
		tp.prevLrnPatterns = nil
		return -1
	}

	// We did find a valid starting point in the past. Now, we need to
	// re-enforce all segments that became active when following this path.

	if tp.params.Verbosity >= 3 {
		fmt.Printf("Discovered path to current input by using start cells from %v steps ago: %v \n",
			numPrevPatterns-startOffset, tp.prevLrnPatterns[startOffset])
	}

	tp.learnBacktrackFrom(startOffset, false)

	// Remove any useless patterns at the head of the input pattern history
	// queue.
	for i := 0; i < numPrevPatterns; i++ {
		if ContainsInt(i, badPatterns) || i <= startOffset {

			if tp.params.Verbosity >= 3 {
				fmt.Printf("Removing useless pattern 1 of: %v from history: %v \n", len(tp.prevLrnPatterns), tp.prevLrnPatterns[0])
			}

			tp.prevLrnPatterns = append(tp.prevLrnPatterns[:0], tp.prevLrnPatterns[1:]...)
		} else {
			break
		}
	}
	return numPrevPatterns - startOffset
}

/*
 Update the learning state. Called from compute() on every iteration
param activeColumns List of active column indices
*/
func (tp *TemporalPooler) updateLearningState(activeColumns []int) {
	// Copy predicted and active states into t-1
	tp.DynamicState.lrnPredictedStateLast = tp.DynamicState.lrnPredictedState.Copy()
	tp.DynamicState.lrnActiveStateLast = tp.DynamicState.lrnActiveState.Copy()

	// Update our learning input history
	if tp.params.MaxLrnBacktrack > 0 {
		if len(tp.prevLrnPatterns) > tp.params.MaxLrnBacktrack {
			tp.prevLrnPatterns = append(tp.prevLrnPatterns[:0], tp.prevLrnPatterns[1:]...)
		}
		tp.prevLrnPatterns = append(tp.prevLrnPatterns, activeColumns)
		if tp.params.Verbosity >= 4 {
			fmt.Printf("Previous learn patterns %v \n", tp.prevLrnPatterns)
		}
	}

	// Process queued up segment updates, now that we have bottom-up, we
	// can update the permanences on the cells that we predicted to turn on
	// and did receive bottom-up
	tp.processSegmentUpdates(activeColumns)

	// Decrement the PAM counter if it is running and increment our learned
	// sequence length
	if tp.pamCounter > 0 {
		tp.pamCounter--
	}
	tp.learnedSeqLength++

	// Phase 1 - turn on the predicted cell in each column that received
	// bottom-up. If there was no predicted cell, pick one to learn to.
	if !tp.resetCalled {
		// Uses lrnActiveState['t-1'] and lrnPredictedState['t-1']
		// computes lrnActiveState['t']
		inSequence := tp.learnPhase1(activeColumns, false)
		// Reset our PAM counter if we are in sequence
		if inSequence {
			tp.pamCounter = tp.params.PamLength
		}
	}

	// Print status of PAM counter, learned sequence length
	if tp.params.Verbosity >= 3 {
		fmt.Printf("pamCounter=%v seqLength=%v \n", tp.pamCounter, tp.learnedSeqLength)
	}

	// Start over on start cells if any of the following occur:
	//    1.) A reset was just called
	//    2.) We have been loo long out of sequence (the pamCounter has expired)
	//    3.) We have reached maximum allowed sequence length.

	//    Note that, unless we are following a reset, we also just learned or
	//    re-enforced connections to the current set of active columns because
	//    this input is still a valid prediction to learn.

	//    It is especially helpful to learn the connections to this input when
	//    you have a maxSeqLength constraint in place. Otherwise, you will have
	//    no continuity at all between sub-sequences of length maxSeqLength.

	if tp.resetCalled || tp.pamCounter == 0 ||
		(tp.params.MaxSeqLength != 0 &&
			tp.learnedSeqLength >= tp.params.MaxSeqLength) {

		if tp.params.Verbosity >= 3 {
			if tp.resetCalled {
				fmt.Printf("Starting over: %v (reset was called) \n", activeColumns)
			} else if tp.pamCounter == 0 {
				fmt.Printf("Starting over: %v (PAM counter expired) \n", activeColumns)
			} else {
				fmt.Printf("Starting over: %v (reached maxSeqLength) \n", activeColumns)
			}
		}

		// Update average learned sequence length - this is a diagnostic statistic
		seqLength := 0
		if tp.pamCounter == 0 {
			seqLength = tp.learnedSeqLength - tp.params.PamLength
		} else {
			seqLength = tp.learnedSeqLength
		}
		if tp.params.Verbosity >= 3 {
			fmt.Printf("learned sequence length was: %v \n", tp.learnedSeqLength)
		}
		tp.updateAvgLearnedSeqLength(float64(seqLength))

		// Backtrack to an earlier starting point, if we find one
		backSteps := 0
		if !tp.resetCalled {
			backSteps = tp.learnBacktrack()
		}

		// Start over in the current time step if reset was called, or we couldn't
		// backtrack.
		if tp.resetCalled || backSteps < 1 {
			tp.DynamicState.lrnActiveState.Clear()
			for _, c := range activeColumns {
				tp.DynamicState.lrnActiveState.Set(c, 0, true)
			}

			// Remove any old input history patterns
			tp.prevLrnPatterns = nil
		}

		// Reset PAM counter
		tp.pamCounter = tp.params.PamLength
		tp.learnedSeqLength = backSteps

		// Clear out any old segment updates from prior sequences
		tp.segmentUpdates = nil

	}

	// Phase 2 - Compute new predicted state. When computing predictions for
	// phase 2, we predict at most one cell per column (the one with the best
	// matching segment).
	tp.learnPhase2(false)

}

/*
 Handle one compute, possibly learning.

param bottomUpInput The bottom-up input, typically from a spatial pooler
param enableLearn If true, perform learning
param computeInfOutput If None, default behavior is to disable the inference
output when enableLearn is on.
If true, compute the inference output
If false, do not compute the inference output

It is an error to have both enableLearn and computeInfOutput set to False

By default, we don't compute the inference output when learning because it
slows things down, but you can override this by passing in True for
computeInfOutput
*/
func (tp *TemporalPooler) Compute(bottomUpInput []bool, enableLearn bool, computeInfOutput bool) []bool {
	if !(enableLearn || computeInfOutput) {
		panic("Enable learn or computeInfOutput must be true")
	}

	// Get the list of columns that have bottom-up
	activeColumns := OnIndices(bottomUpInput)

	if enableLearn {
		tp.lrnIterationIdx++
	}
	tp.iterationIdx++

	if tp.params.Verbosity >= 3 {
		fmt.Printf("\n==== Iteration: %v \n ", tp.iterationIdx)
		fmt.Println("Active cols:", activeColumns)
	}

	// Update segment duty cycles if we are crossing a "tier"
	// We determine if it's time to update the segment duty cycles. Since the
	// duty cycle calculation is a moving average based on a tiered alpha, it is
	// important that we update all segments on each tier boundary

	if enableLearn {
		if ContainsInt(tp.lrnIterationIdx, SegmentDutyCycleTiers) {

			for c := 0; c < tp.params.NumberOfCols; c++ {
				for i := 0; i < tp.params.CellsPerColumn; i++ {
					for _, segment := range tp.cells[c][i] {
						segment.dutyCycle(false, false)
					}
				}
			}

		}

	}

	// Update the average input density
	if tp.avgInputDensity == -1 {
		tp.avgInputDensity = float64(len(activeColumns))
	} else {
		tp.avgInputDensity = (0.99*tp.avgInputDensity + 0.01*float64(len(activeColumns)))
	}

	// First, update the inference state
	// As a speed optimization for now (until we need online learning), skip
	// computing the inference output while learning
	if computeInfOutput {
		tp.updateInferenceState(activeColumns)
	}

	// Next, update the learning state
	if enableLearn {
		tp.updateLearningState(activeColumns)

		// Apply global decay, and remove synapses and/or segments.
		// Synapses are removed if their permanence value is <= 0.
		// Segments are removed when they don't have synapses anymore.
		// Removal of synapses can trigger removal of whole segments!
		// todo: isolate the synapse/segment retraction logic so that
		// it can be called in adaptSegments, in the case where we
		// do global decay only episodically.

		if tp.params.GlobalDecay > 0.0 && (tp.lrnIterationIdx%tp.params.MaxAge) == 0 {
			for _, col := range tp.cells {
				for _, cell := range col {

					si := 0 //New end of segment slice
					for _, segment := range cell {
						age := tp.lrnIterationIdx - segment.lastActiveIteration
						if age <= tp.params.MaxAge {
							//keep segment but skip syn trimming
							cell[si] = segment
							si++
							continue
						}

						w := 0
						for _, synapse := range segment.syns {
							// decrease permanence
							synapse.Permanence = synapse.Permanence - float64(tp.params.GlobalDecay)
							//remove synapse if permanence too low
							if synapse.Permanence <= 0 {
								continue
							}
							//otherwise keep synapse and increment synapse end of slice
							segment.syns[w] = synapse
							w++
						}
						segment.syns = segment.syns[:w]

						//trim segment if no synapses remaining
						if len(segment.syns) == 0 {
							continue
						}
						//otherwise keep segment increment end  of slice counter
						cell[si] = segment
						si++
					}

				}
			}

		} // end globalDecay if

		// Teach the trivial predictors
		if tp.trivialPredictor != nil {
			tp.trivialPredictor.learn(activeColumns)
		}

		// Update the prediction score stats
		// Learning always includes inference
		var predictedState *SparseBinaryMatrix
		if tp.params.CollectStats {
			if computeInfOutput {
				predictedState = tp.DynamicState.infPredictedStateLast.Copy()
			} else {
				predictedState = tp.DynamicState.lrnPredictedStateLast.Copy()
			}

			tp.updateStatsInferEnd(tp.internalStats, activeColumns,
				predictedState, tp.DynamicState.colConfidenceLast)

			// Make trivial predictions and collect stats

			if tp.trivialPredictor != nil {
				for _, method := range tp.trivialPredictor.Methods {
					if computeInfOutput {
						tp.trivialPredictor.infer(activeColumns)
					}

					temp := NewSparseBinaryMatrixFromDense([][]bool{tp.trivialPredictor.State[method].PredictedStateLast})

					tp.updateStatsInferEnd(tp.trivialPredictor.InternalStats[method],
						activeColumns,
						temp,
						tp.trivialPredictor.State[method].ConfidenceLast)
				}
			}

		}

	}

	// Finally return the TP output
	result := tp.computeOutput()

	// Print diagnostic information based on the current verbosity level
	//tp.printComputeEnd(result, enableLearn)

	tp.resetCalled = false
	return result

}
