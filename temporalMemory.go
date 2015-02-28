package htm

import (
	//"fmt"
	"github.com/cznic/mathutil"
	// 	"github.com/zacg/floats"
	// 	"github.com/zacg/go.matrix"
	"github.com/nupic-community/htm/utils"
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

//Create default temporal memory params
func NewTemporalMemoryParams() *TemporalMemoryParams {
	p := new(TemporalMemoryParams)

	p.ColumnDimensions = []int{2048}
	p.CellsPerColumn = 32
	p.ActivationThreshold = 13
	p.LearningRadius = 2048
	p.InitialPermanence = 0.21
	p.ConnectedPermanence = 0.50
	p.MinThreshold = 10
	p.MaxNewSynapseCount = 20
	p.PermanenceIncrement = 0.10
	p.PermanenceDecrement = 0.10
	p.Seed = 42

	return p
}

/*
Temporal memory
*/
type TemporalMemory struct {
	params                   *TemporalMemoryParams
	ActiveCells              []int
	PredictiveCells          []int
	ActiveSegments           []int
	ActiveSynapsesForSegment map[int][]int
	WinnerCells              []int
	Connections              *TemporalMemoryConnections
}

//Create new temporal memory
func NewTemporalMemory(params *TemporalMemoryParams) *TemporalMemory {
	tm := new(TemporalMemory)
	tm.params = params
	tm.Connections = NewTemporalMemoryConnections(params.MaxNewSynapseCount,
		params.CellsPerColumn, params.ColumnDimensions)
	//TODO: refactor into encapsulated RNG
	rand.Seed(int64(params.Seed))
	return tm
}

//Feeds input record through TM, performing inference and learning.
//Updates member variables with new state.
func (tm *TemporalMemory) Compute(activeColumns []int, learn bool) {

	activeCells, winnerCells, activeSynapsesForSegment, activeSegments, predictiveCells := tm.computeFn(activeColumns,
		tm.PredictiveCells,
		tm.ActiveSegments,
		tm.ActiveSynapsesForSegment,
		tm.WinnerCells,
		tm.Connections,
		learn)

	tm.ActiveCells = activeCells
	tm.WinnerCells = winnerCells
	tm.ActiveSynapsesForSegment = activeSynapsesForSegment
	tm.ActiveSegments = activeSegments
	tm.PredictiveCells = predictiveCells

}

// helper for compute().
//Returns new state
func (tm *TemporalMemory) computeFn(activeColumns []int,
	prevPredictiveCells []int,
	prevActiveSegments []int,
	prevActiveSynapsesForSegment map[int][]int,
	prevWinnerCells []int,
	connections *TemporalMemoryConnections,
	learn bool) (activeCells []int,
	winnerCells []int,
	activeSynapsesForSegment map[int][]int,
	activeSegments []int,
	predictiveCells []int) {

	var predictedColumns []int

	activeCells, winnerCells, predictedColumns = tm.activateCorrectlyPredictiveCells(
		prevPredictiveCells,
		activeColumns,
		connections)

	_activeCells, _winnerCells, learningSegments := tm.burstColumns(activeColumns,
		predictedColumns,
		prevActiveSynapsesForSegment,
		connections)

	utils.Add(activeCells, _activeCells)
	utils.Add(winnerCells, _winnerCells)

	if learn {
		tm.learnOnSegments(prevActiveSegments,
			learningSegments,
			prevActiveSynapsesForSegment,
			winnerCells,
			prevWinnerCells,
			connections)
	}

	activeSynapsesForSegment = tm.computeActiveSynapses(activeCells, connections)

	activeSegments, predictiveCells = tm.computePredictiveCells(activeSynapsesForSegment,
		connections)

	return activeCells,
		winnerCells,
		activeSynapsesForSegment,
		activeSegments,
		predictiveCells

}

//Indicates the start of a new sequence. Resets sequence state of the TM.
func (tm *TemporalMemory) Reset() {
	tm.ActiveCells = tm.ActiveCells[:0]
	tm.PredictiveCells = tm.PredictiveCells[:0]
	tm.ActiveSegments = tm.ActiveSegments[:0]
	tm.WinnerCells = tm.WinnerCells[:0]
}

/*
Phase 1: Activate the correctly predictive cells.
Pseudocode:
- for each prev predictive cell
- if in active column
- mark it as active
- mark it as winner cell
- mark column as predicted
*/
func (tm *TemporalMemory) activateCorrectlyPredictiveCells(prevPredictiveCells []int,
	activeColumns []int,
	connections *TemporalMemoryConnections) (activeCells []int,
	winnerCells []int,
	predictedColumns []int) {

	for _, cell := range prevPredictiveCells {
		column := connections.ColumnForCell(cell)
		if utils.ContainsInt(column, activeColumns) {
			activeCells = append(activeCells, cell)
			winnerCells = append(winnerCells, cell)
			//TODO: change this to a set data structure
			if !utils.ContainsInt(column, predictedColumns) {
				predictedColumns = append(predictedColumns, column)
			}
		}
	}

	return activeCells, winnerCells, predictedColumns
}

/*
Phase 2: Burst unpredicted columns.
Pseudocode:
- for each unpredicted active column
- mark all cells as active
- mark the best matching cell as winner cell
- (learning)
- if it has no matching segment
- (optimization) if there are prev winner cells
- add a segment to it
- mark the segment as learning
*/
func (tm *TemporalMemory) burstColumns(activeColumns []int,
	predictedColumns []int,
	prevActiveSynapsesForSegment map[int][]int,
	connections *TemporalMemoryConnections) (activeCells []int,
	winnerCells []int,
	learningSegments []int) {

	unpredictedColumns := utils.Complement(activeColumns, predictedColumns)

	for _, column := range unpredictedColumns {
		cells := connections.CellsForColumn(column)
		activeCells = utils.Add(activeCells, cells)

		bestCell, bestSegment := tm.getBestMatchingCell(column,
			prevActiveSynapsesForSegment,
			connections)

		winnerCells = append(winnerCells, bestCell)

		if bestSegment == -1 {
			//TODO: (optimization) Only do this if there are prev winner cells
			bestSegment = connections.CreateSegment(bestCell)
		}
		//TODO: change to set data structure
		if !utils.ContainsInt(bestSegment, learningSegments) {
			learningSegments = append(learningSegments, bestSegment)
		}
	}

	return activeCells, winnerCells, learningSegments
}

/*
Phase 3: Perform learning by adapting segments.
Pseudocode:
- (learning) for each prev active or learning segment
- if learning segment or from winner cell
- strengthen active synapses
- weaken inactive synapses
- if learning segment
- add some synapses to the segment
- subsample from prev winner cells
*/
func (tm *TemporalMemory) learnOnSegments(prevActiveSegments []int,
	learningSegments []int,
	prevActiveSynapsesForSegment map[int][]int,
	winnerCells []int,
	prevWinnerCells []int,
	connections *TemporalMemoryConnections) {

	tm.lrnOnSegments(prevActiveSegments, false, prevActiveSynapsesForSegment, winnerCells, prevWinnerCells, connections)
	tm.lrnOnSegments(learningSegments, true, prevActiveSynapsesForSegment, winnerCells, prevWinnerCells, connections)

}

//helper
func (tm *TemporalMemory) lrnOnSegments(segments []int,
	isLearningSegments bool,
	prevActiveSynapsesForSegment map[int][]int,
	winnerCells []int,
	prevWinnerCells []int,
	connections *TemporalMemoryConnections) {

	for _, segment := range segments {
		isFromWinnerCell := utils.ContainsInt(connections.CellForSegment(segment), winnerCells)
		activeSynapses := tm.getConnectedActiveSynapsesForSegment(segment,
			prevActiveSynapsesForSegment,
			0,
			connections)

		if isLearningSegments || isFromWinnerCell {
			tm.adaptSegment(segment, activeSynapses, connections)
		}

		if isLearningSegments {
			n := tm.params.MaxNewSynapseCount - len(activeSynapses)
			for _, sourceCell := range tm.pickCellsToLearnOn(n,
				segment,
				winnerCells,
				connections) {
				connections.CreateSynapse(segment, sourceCell, tm.params.InitialPermanence)
			}
		}

	}

}

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
		for _, synapse := range connections.SynapsesForSourceCell(cell) {
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
	bestCell = -1
	bestSegment = -1

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

	if bestCell == -1 {
		bestCell = tm.getLeastUsedCell(column, connections)
	}

	return bestCell, bestSegment
}

// Gets the segment on a cell with the largest number of activate synapses,
// including all synapses with non-zero permanences.
func (tm *TemporalMemory) getBestMatchingSegment(cell int, activeSynapsesForSegment map[int][]int,
	connections *TemporalMemoryConnections) (bestSegment int, connectedActiveSynapses []int) {
	maxSynapses := tm.params.MinThreshold
	bestSegment = -1

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
			perm -= tm.params.PermanenceDecrement
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
