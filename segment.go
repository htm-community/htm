package htm

import (
	//"fmt"
	"github.com/cznic/mathutil"
	//"github.com/skelterjohn/go.matrix"
	"math"
	//"math/rand"
	//"sort"
	"github.com/gonum/floats"
)

type Synapse struct {
	SrcCellCol int
	SrcCellIdx int
	Permanence float64
}

// The Segment struct is a container for all of the segment variables and
//the synapses it owns.
type Segment struct {
	tp                        *TemporalPooler
	segId                     int
	isSequenceSeg             bool
	lastActiveIteration       int
	positiveActivations       int
	totalActivations          int
	lastPosDutyCycle          float64
	lastPosDutyCycleIteration int
	syns                      []Synapse
	dutyCycleTiers            []int
	dutyCycleAlphas           []float64
}

func (s *Segment) Equals(seg *Segment) bool {
	synsEqual := true

	if len(s.syns) != len(seg.syns) {
		return false
	}

	for idx, val := range s.syns {
		if seg.syns[idx].Permanence != val.Permanence ||
			seg.syns[idx].SrcCellCol != val.SrcCellCol || seg.syns[idx].SrcCellIdx != val.SrcCellIdx {
			return false
		}
	}

	if len(s.dutyCycleTiers) != len(seg.dutyCycleTiers) {
		return false
	}

	for idx, val := range s.dutyCycleTiers {
		if seg.dutyCycleTiers[idx] != val {
			return false
		}
	}

	if len(s.dutyCycleAlphas) != len(seg.dutyCycleAlphas) {
		return false
	}

	for idx, val := range s.dutyCycleAlphas {
		if seg.dutyCycleAlphas[idx] != val {
			return false
		}
	}

	return synsEqual &&
		s.tp == seg.tp &&
		s.segId == seg.segId &&
		s.isSequenceSeg == seg.isSequenceSeg &&
		s.lastActiveIteration == seg.lastActiveIteration &&
		s.positiveActivations == seg.positiveActivations &&
		s.totalActivations == seg.totalActivations &&
		s.lastPosDutyCycle == seg.lastPosDutyCycle &&
		s.lastPosDutyCycleIteration == seg.lastPosDutyCycleIteration

}

//Creates a new segment
func NewSegment(tp *TemporalPooler, isSequenceSeg bool) *Segment {
	seg := Segment{}
	seg.tp = tp
	seg.segId = tp.GetSegId()
	seg.isSequenceSeg = isSequenceSeg
	seg.lastActiveIteration = tp.lrnIterationIdx
	seg.positiveActivations = 1
	seg.totalActivations = 1

	seg.lastPosDutyCycle = 1.0 / float64(tp.lrnIterationIdx)
	seg.lastPosDutyCycleIteration = tp.lrnIterationIdx

	seg.dutyCycleTiers = []int{0, 100, 320, 1000,
		3200, 10000, 32000, 100000, 320000}

	seg.dutyCycleAlphas = []float64{0, 0.0032, 0.0010, 0.00032,
		0.00010, 0.000032, 0.00001, 0.0000032,
		0.0000010}

	//TODO: initialize synapse collection

	return &seg
}

// func (s *Segment) Equal(other *Segment) bool {
// }

/*
Compute/update and return the positive activations duty cycle of
this segment. This is a measure of how often this segment is
providing good predictions.

param active True if segment just provided a good prediction
param readOnly If True, compute the updated duty cycle, but don't change
the cached value. This is used by debugging print statements.

returns The duty cycle, a measure of how often this segment is
providing good predictions.

**NOTE:** This method relies on different schemes to compute the duty cycle
based on how much history we have. In order to support this tiered
approach **IT MUST BE CALLED ON EVERY SEGMENT AT EACH DUTY CYCLE TIER**
(ref dutyCycleTiers).

When we don't have a lot of history yet (first tier), we simply return
number of positive activations / total number of iterations

After a certain number of iterations have accumulated, it converts into
a moving average calculation, which is updated only when requested
since it can be a bit expensive to compute on every iteration (it uses
the pow() function).

The duty cycle is computed as follows:

dc[t] = (1-alpha) * dc[t-1] + alpha * value[t]

If the value[t] has been 0 for a number of steps in a row, you can apply
all of the updates at once using:

dc[t] = (1-alpha)^(t-lastT) * dc[lastT]

We use the alphas and tiers as defined in ref dutyCycleAlphas and
ref dutyCycleTiers.
*/
func (s *Segment) dutyCycle(active, readOnly bool) float64 {

	// For tier #0, compute it from total number of positive activations seen
	if s.tp.lrnIterationIdx <= s.dutyCycleTiers[1] {
		dutyCycle := float64(s.positiveActivations) / float64(s.tp.lrnIterationIdx)
		if !readOnly {
			s.lastPosDutyCycleIteration = s.tp.lrnIterationIdx
			s.lastPosDutyCycle = dutyCycle
		}
		return dutyCycle
	}

	// How old is our update?
	age := s.tp.lrnIterationIdx - s.lastPosDutyCycleIteration

	//If it's already up to date, we can returned our cached value.
	if age == 0 && !active {
		return s.lastPosDutyCycle
	}

	alpha := 0.0
	//Figure out which alpha we're using
	for i := len(s.dutyCycleTiers); i > 0; i-- {
		if s.tp.lrnIterationIdx > s.dutyCycleTiers[i] {
			alpha = s.dutyCycleAlphas[i]
			break
		}
	}

	// Update duty cycle
	dutyCycle := math.Pow(1.0-alpha, float64(age)) * s.lastPosDutyCycle

	if active {
		dutyCycle += alpha
	}

	// Update cached values if not read-only
	if !readOnly {
		s.lastPosDutyCycleIteration = s.tp.lrnIterationIdx
		s.lastPosDutyCycle = dutyCycle
	}

	return dutyCycle
}

/*
Free up some synapses in this segment. We always free up inactive
synapses (lowest permanence freed up first) before we start to free up
active ones.

@param numToFree number of synapses to free up
@param inactiveSynapseIndices list of the inactive synapse indices.
*/

func (s *Segment) freeNSynapses(numToFree int, inactiveSynapseIndices []int) {
	//Make sure numToFree isn't larger than the total number of syns we have
	if numToFree > len(s.syns) {
		panic("Number to free cannot be larger than existing synapses.")
	}

	var candidates []int
	// Remove the lowest perm inactive synapses first
	if len(inactiveSynapseIndices) > 0 {
		perms := make([]float64, len(inactiveSynapseIndices))
		for idx, _ := range perms {
			perms[idx] = s.syns[idx].Permanence
		}
		var indexes []int
		floats.Argsort(perms, indexes)
		//sort perms
		cSize := mathutil.Min(numToFree, len(perms))
		candidates = make([]int, cSize)
		//indexes[0:cSize]
		for i := 0; i < cSize; i++ {
			candidates[i] = inactiveSynapseIndices[indexes[i]]
		}
	}

	// Do we need more? if so, remove the lowest perm active synapses too
	var activeSynIndices []int
	if len(candidates) < numToFree {
		for i := 0; i < len(s.syns); i++ {
			if !ContainsInt(i, inactiveSynapseIndices) {
				activeSynIndices = append(activeSynIndices, i)
			}
		}

		perms := make([]float64, len(activeSynIndices))
		for i := range perms {
			perms[i] = s.syns[i].Permanence
		}
		var indexes []int
		floats.Argsort(perms, indexes)

		moreToFree := numToFree - len(candidates)
		//moreCandidates := make([]int, moreToFree)
		for i := 0; i < moreToFree; i++ {
			candidates = append(candidates, activeSynIndices[indexes[i]])
		}
	}

	// Delete candidate syns by copying undeleted to new slice
	var newSyns []Synapse
	for idx, val := range s.syns {
		if !ContainsInt(idx, candidates) {
			newSyns = append(newSyns, val)
		}
	}
	s.syns = newSyns

}

/*
Update a set of synapses in the segment.

param synapses List of synapse indices to update
param delta How much to add to each permanence

returns True if synapse reached 0
*/

func (s *Segment) updateSynapses(synapses []int, delta float64) bool {
	hitZero := false

	if delta > 0 {
		for idx, _ := range synapses {
			s.syns[idx].Permanence += delta
			// Cap synapse permanence at permanenceMax
			if s.syns[idx].Permanence > s.tp.params.PermanenceMax {
				s.syns[idx].Permanence = s.tp.params.PermanenceMax
			}
		}
	} else {
		for idx, _ := range synapses {
			s.syns[idx].Permanence += delta
			// Cap min synapse permanence to 0 in case there is no global decay
			if s.syns[idx].Permanence <= 0 {
				s.syns[idx].Permanence = 0
				hitZero = true
			}
		}
	}

	return hitZero
}

/*
Adds a new synapse
*/

func (s *Segment) AddSynapse(srcCellCol, srcCellIdx int, perm float64) {
	s.syns = append(s.syns, Synapse{srcCellCol, srcCellIdx, perm})
}

/*
 Return a segmentUpdate data structure containing a list of proposed
changes to segment s. Let activeSynapses be the list of active synapses
where the originating cells have their activeState output = true at time step
t. (This list is empty if s is None since the segment doesn't exist.)
newSynapses is an optional argument that defaults to false. If newSynapses
is true, then newSynapseCount - len(activeSynapses) synapses are added to
activeSynapses. These synapses are randomly chosen from the set of cells
that have learnState = true at timeStep.
*/

func (tp *TemporalPooler) getSegmentActiveSynapses(c int, i int, s *Segment, activeState *SparseBinaryMatrix, newSynapses bool) *SegmentUpdate {
	var activeSynapses []SynapseUpdateState

	if s != nil {
		for idx, val := range s.syns {
			if activeState.Get(val.SrcCellIdx, val.SrcCellCol) {
				temp := SynapseUpdateState{}
				temp.Index = idx
				activeSynapses = append(activeSynapses, temp)
			}
		}
	}

	if newSynapses {
		nSynapsesToAdd := tp.params.NewSynapseCount - len(activeSynapses)
		newSyns := tp.chooseCellsToLearnFrom(s, nSynapsesToAdd, activeState)

		for _, val := range newSyns {
			temp := SynapseUpdateState{}
			temp.Index = val.Row
			temp.CellIndex = val.Col
			temp.New = true
			activeSynapses = append(activeSynapses, temp)
		}
	}

	// It's still possible that activeSynapses is empty, and this will
	// be handled in addToSegmentUpdates
	result := new(SegmentUpdate)
	result.activeSynapses = activeSynapses
	result.columnIdx = c
	result.cellIdx = i
	result.segment = s
	return result

}
