package htm

import (
	// 	"fmt"
	"github.com/cznic/mathutil"
	// 	"github.com/zacg/floats"
	// 	"github.com/zacg/go.matrix"
	"github.com/zacg/htm/utils"
	//"github.com/zacg/ints"
	"math"
	"math/rand"
	// 	//"sort"
)

/*
Params for intializing temporal memory
*/
type TemporalMemoryParams struct {
	//Column dimensions
	ColumnDimensions []int
	CellsPerColumn   int
	//If the number of active connected synapses on a segment is at least
	//this threshold, the segment is said to be active.
	ActivationThreshold int
	//Radius around cell from which it can sample to form distal dendrite
	//connections.
	LearningRadius    int
	InitialPermanence float64
	//If the permanence value for a synapse is greater than this value, it is said
	//to be connected.
	ConnectedPermanence float64
	//If the number of synapses active on a segment is at least this threshold,
	//it is selected as the best matching cell in a bursing column.
	MinThreshold int
	//The maximum number of synapses added to a segment during learning.
	MaxNewSynapseCount  int
	PermanenceIncrement float64
	PermanenceDecrement float64
	//rand seed
	Seed int
}

/*
Temporal memory
*/
type TemporalMemory struct {
	params *TemporalMemoryParams
}

//Create new temporal memory
func NewTemporalMemory(params *TemporalMemoryParams) *TemporalMemory {
	tm := new(TemporalMemory)
	tm.params = params
	return tm
}

//Feeds input record through TM, performing inference and learning.
//Updates member variables with new state.
// func (tm *TemporalMemory) Compute(activeColumns []int, learn bool) {

// }

// func compute() {

// }

/*
 Phase 4: Compute predictive cells due to lateral input
on distal dendrites.

Pseudocode:

- for each distal dendrite segment with activity >= activationThreshold
- mark the segment as active
- mark the cell as predictive
*/
func (tm *TemporalMemory) computePredictiveCells(activeSynapsesForSegment map[int][]int,
	connections *TemporalMemoryConnections) (activeSegments []int, predictiveCells []int) {

	for segment, _ := range activeSynapsesForSegment {
		synapses := tm.getConnectedActiveSynapsesForSegment(segment,
			activeSynapsesForSegment,
			tm.params.ConnectedPermanence,
			connections)
		if len(synapses) >= tm.params.ActivationThreshold {
			activeSegments = append(activeSegments, segment)
			predictiveCells = append(predictiveCells, connections.CellForSegment(segment))
		}
	}

	return activeSegments, predictiveCells
}

// Forward propagates activity from active cells to the synapses that touch
// them, to determine which synapses are active.
func (tm *TemporalMemory) computeActiveSynapses(activeCells []int,
	connections *TemporalMemoryConnections) map[int][]int {

	activeSynapsesForSegment := make(map[int][]int)

	for _, cell := range activeCells {
		for synapse := range connections.SynapsesForSourceCell(cell) {
			segment := connections.DataForSynapse(synapse).Segment
			activeSynapsesForSegment[segment] = append(activeSynapsesForSegment[segment], synapse)
		}
	}

	return activeSynapsesForSegment
}

// Gets the cell with the best matching segment
//(see `TM.getBestMatchingSegment`) that has the largest number of active
//synapses of all best matching segments.
//If none were found, pick the least used cell (see `TM.getLeastUsedCell`).
func (tm *TemporalMemory) getBestMatchingCell(column int, activeSynapsesForSegment map[int][]int,
	connections *TemporalMemoryConnections) (bestCell int, bestSegment int) {

	maxSynapses := 0
	cells := connections.CellsForColumn(column)

	for _, cell := range cells {
		segment, connectedActiveSynapses := tm.getBestMatchingSegment(cell,
			activeSynapsesForSegment,
			connections)

		if segment > -1 && len(connectedActiveSynapses) > maxSynapses {
			maxSynapses = len(connectedActiveSynapses)
			bestCell = cell
			bestSegment = segment
		}
	}

	if bestCell == 0 {
		bestCell = tm.getLeastUsedCell(column, connections)
	}

	return bestCell, bestSegment
}

// Gets the segment on a cell with the largest number of activate synapses,
// including all synapses with non-zero permanences.
func (tm *TemporalMemory) getBestMatchingSegment(cell int, activeSynapsesForSegment map[int][]int,
	connections *TemporalMemoryConnections) (bestSegment int, connectedActiveSynapses []int) {

	maxSynapses := tm.params.MinThreshold

	for _, segment := range connections.SegmentsForCell(cell) {
		synapses := tm.getConnectedActiveSynapsesForSegment(segment,
			activeSynapsesForSegment,
			0,
			connections)

		if len(synapses) >= maxSynapses {
			maxSynapses = len(synapses)
			bestSegment = segment
			connectedActiveSynapses = synapses
		}

	}

	return bestSegment, connectedActiveSynapses
}

// Gets the cell with the smallest number of segments.
// Break ties randomly.
func (tm *TemporalMemory) getLeastUsedCell(column int, connections *TemporalMemoryConnections) int {
	cells := connections.CellsForColumn(column)
	leastUsedCells := make([]int, 0, len(cells))
	minNumSegments := math.MaxInt64

	for _, cell := range cells {
		numSegments := len(connections.SegmentsForCell(cell))

		if numSegments < minNumSegments {
			minNumSegments = numSegments
			leastUsedCells = leastUsedCells[:0]
		}

		if numSegments == minNumSegments {
			leastUsedCells = append(leastUsedCells, cell)
		}
	}

	//pick random cell
	return leastUsedCells[rand.Intn(len(leastUsedCells))]
}

//Returns the synapses on a segment that are active due to lateral input
//from active cells.
func (tm *TemporalMemory) getConnectedActiveSynapsesForSegment(segment int,
	activeSynapsesForSegment map[int][]int, permanenceThreshold float64, connections *TemporalMemoryConnections) []int {

	if _, ok := activeSynapsesForSegment[segment]; !ok {
		return []int{}
	}

	connectedSynapses := make([]int, 0, len(activeSynapsesForSegment))

	//TODO: (optimization) Can skip this logic if permanenceThreshold = 0
	for _, synIdx := range activeSynapsesForSegment[segment] {
		perm := connections.DataForSynapse(synIdx).Permanence
		if perm >= permanenceThreshold {
			connectedSynapses = append(connectedSynapses, synIdx)
		}
	}

	return connectedSynapses
}

// Updates synapses on segment.
// Strengthens active synapses; weakens inactive synapses.
func (tm *TemporalMemory) adaptSegment(segment int, activeSynapses []int,
	connections *TemporalMemoryConnections) {

	for _, synIdx := range connections.SynapsesForSegment(segment) {
		syn := connections.DataForSynapse(synIdx)
		perm := syn.Permanence

		if utils.ContainsInt(synIdx, activeSynapses) {
			perm += tm.params.PermanenceIncrement
		} else {
			perm += tm.params.PermanenceDecrement
		}
		//enforce min/max bounds
		perm = math.Max(0.0, math.Min(1.0, perm))
		connections.UpdateSynapsePermanence(synIdx, perm)
	}

}

//Pick cells to form distal connections to.
func (tm *TemporalMemory) pickCellsToLearnOn(n int, segment int,
	winnerCells []int, connections *TemporalMemoryConnections) []int {

	candidates := make([]int, len(winnerCells))
	copy(candidates, winnerCells)

	for _, val := range connections.SynapsesForSegment(segment) {
		syn := connections.DataForSynapse(val)
		for idx, val := range candidates {
			if val == syn.SourceCell {
				candidates = append(candidates[:idx], candidates[idx+1:]...)
				break
			}
		}
	}

	//Shuffle candidates
	for i := range candidates {
		j := rand.Intn(i + 1)
		candidates[i], candidates[j] = candidates[j], candidates[i]
	}

	n = mathutil.Min(n, len(candidates))
	return candidates[:n]
}
