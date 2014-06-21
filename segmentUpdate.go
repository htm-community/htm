package htm

import (
//"fmt"
//"github.com/cznic/mathutil"
//"github.com/skelterjohn/go.matrix"
//"math"
//"math/rand"
//"sort"
)

type SynapseUpdateState struct {
	New       bool
	Index     int
	CellIndex int //only set when new
}

type SegmentUpdate struct {
	columnIdx        int
	cellIdx          int
	segment          *Segment
	activeSynapses   []SynapseUpdateState
	sequenceSegment  bool
	phase1Flag       bool
	weaklyPredicting bool
	lrnIterationIdx  int
}

type UpdateState struct {
	//creationdate refers to iteration idx
	CreationDate int
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
 This function applies segment update information to a segment in a
cell.

Synapses on the active list get their permanence counts incremented by
permanenceInc. All other synapses get their permanence counts decremented
by permanenceDec.

We also increment the positiveActivations count of the segment.

param segUpdate SegmentUpdate instance
returns True if some synapses were decremented to 0 and the segment is a
candidate for trimming
*/

func (segUpdate *SegmentUpdate) adaptSegments(tp *TemporalPooler) bool {
	// This will be set to True if detect that any syapses were decremented to 0
	trimSegment := false

	// segUpdate.segment is None when creating a new segment
	c, i, segment := segUpdate.columnIdx, segUpdate.cellIdx, segUpdate.segment

	// update.activeSynapses can be empty.
	// If not, it can contain either or both integers and tuples.
	// The integers are indices of synapses to update.
	// The tuples represent new synapses to create (src col, src cell in col).
	// We pre-process to separate these various element types.
	// synToCreate is not empty only if positiveReinforcement is True.
	// NOTE: the synapse indices start at *1* to skip the segment flags.
	activeSynapses := segUpdate.activeSynapses

	var synToUpdate []int
	for _, val := range activeSynapses {
		if !val.New {
			synToUpdate = append(synToUpdate, val.Index)
		}
	}

	if segment != nil {
		//modify existing segment
		// Mark it as recently useful
		segment.lastActiveIteration = tp.lrnIterationIdx

		// Update frequency and positiveActivations
		segment.positiveActivations++ // positiveActivations += 1
		segment.dutyCycle(true, false)

		// First, decrement synapses that are not active
		lastSynIndex := len(segment.syns) - 1

		var inactiveSynIndices []int
		for i := 0; i < lastSynIndex+1; i++ {
			if !ContainsInt(i, synToUpdate) {
				inactiveSynIndices = append(inactiveSynIndices, i)
			}
		}

		trimSegment = segment.updateSynapses(inactiveSynIndices, -tp.params.PermanenceDec)

		// Now, increment active synapses
		var activeSynIndices []int
		for _, val := range activeSynapses {
			if val.Index <= lastSynIndex {
				activeSynIndices = append(activeSynIndices, val.Index)
			}
		}

		segment.updateSynapses(activeSynIndices, tp.params.PermanenceInc)

		// Finally, create new synapses if needed
		var synsToAdd []SynapseUpdateState
		for _, val := range activeSynapses {
			if val.New {
				synsToAdd = append(synsToAdd, val)
			}
		}

		// If we have fixed resources, get rid of some old syns if necessary
		if tp.params.MaxSynapsesPerSegment > 0 && len(synsToAdd)+len(segment.syns) > tp.params.MaxSynapsesPerSegment {
			numToFree := (len(segment.syns) + len(synsToAdd)) - tp.params.MaxSynapsesPerSegment
			segment.freeNSynapses(numToFree, inactiveSynIndices)
		}

		for _, val := range synsToAdd {
			segment.AddSynapse(val.Index, val.CellIndex, tp.params.InitialPerm)
		}

	} else {
		//create new segment
		newSegment := new(Segment)
		newSegment.tp = tp
		newSegment.isSequenceSeg = segUpdate.sequenceSegment

		for _, val := range activeSynapses {
			newSegment.AddSynapse(val.Index, val.CellIndex, tp.params.InitialPerm)
		}

		tp.cells[c][i] = append(tp.cells[c][i], *newSegment)
	}

	return trimSegment
}
