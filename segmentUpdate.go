package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/skelterjohn/go.matrix"
	"math"
	"math/rand"
	"sort"
)

type SegmentUpdate struct {
	columnIdx        int
	cellIdx          int
	segment          *Segment
	activeSynapses   []int
	sequenceSegment  bool
	phase1Flag       bool
	weaklyPredicting bool
	lrnIterationIdx  int
}

type UpdateState struct {
	IterationIdx int
	Update       *SegmentUpdate
}

/*
 Store a dated potential segment update. The "date" (iteration index) is used
later to determine whether the update is too old and should be forgotten.
This is controlled by parameter segUpdateValidDuration.
*/

func (su *TemporalPooler) addToSegmentUpdates(c, i int, segUpdate *SegmentUpdate) {
	if segUpdate == nil || len(segUpdate.activeSynapses) == 0 {
		return
	}

	// key = (column index, cell index in column)
	key := TupleInt{c, i}

	newUpdate := UpdateState{su.lrnIterationIdx, segUpdate}
	if _, ok := su.segmentUpdates[key]; ok {
		su.segmentUpdates[key] = append(su.segmentUpdates[key], newUpdate)
	} else {
		su.segmentUpdates[key] = []UpdateState{newUpdate}
	}

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
	return su.colConfidence
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

func (su *TemporalPooler) predict(nSteps int) {
	// Save the TP dynamic state, we will use to revert back in the end
    pristineTPDynamicState = self._getTPDynamicState()

    //assert (nSteps>0)
    if(nSteps <= 0){
    	panic("nSteps must be greater than zero")
    }

    // multiStepColumnPredictions holds all the future prediction.
    //multiStepColumnPredictions = numpy.zeros((nSteps, self.numberOfCols), dtype="float32")
    multiStepColumnPredictions := matrix.MakeDenseMatrix(elements, nSteps, su.params.NumberOfCols)

    // This is a (nSteps-1)+half loop. Phase 2 in both learn and infer methods
    // already predicts for timestep (t+1). We use that prediction for free and
    // save the half-a-loop of work.

    step = 0
    while True:
      // We get the prediction for the columns in the next time step from
      // the topDownCompute method. It internally uses confidences.
      multiStepColumnPredictions[step, :] = self.topDownCompute()

      // Cleanest way in python to handle one and half loops
      if step == nSteps-1:
        break
      step += 1

      // Copy t-1 into t
      self.infActiveState['t-1'][:, :] = self.infActiveState['t'][:, :]
      self.infPredictedState['t-1'][:, :] = self.infPredictedState['t'][:, :]
      self.cellConfidence['t-1'][:, :] = self.cellConfidence['t'][:, :]

      // Predicted state at "t-1" becomes the active state at "t"
      self.infActiveState['t'][:, :] = self.infPredictedState['t-1'][:, :]

      // Predicted state and confidence are set in phase2.
      self.infPredictedState['t'].fill(0)
      self.cellConfidence['t'].fill(0.0)
      self.inferPhase2()

    // Revert the dynamic state to the saved state
    self._setTPDynamicState(pristineTPDynamicState)

    return multiStepColumnPredictions


}
