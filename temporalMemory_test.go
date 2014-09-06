package htm

import (
	//"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestPickCellsToLearnOnAvoidDuplicates(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tmp.MaxNewSynapseCount = 1000
	tm := NewTemporalMemory(tmp)

	connections := tm.Connections
	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)

	winnerCells := []int{233, 144}

	// Ensure that no additional (duplicate) cells were picked
	assert.Equal(t, winnerCells, tm.pickCellsToLearnOn(2, 0, winnerCells, connections))

}

func TestPickCellsToLearnOn(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections
	connections.CreateSegment(0)

	winnerCells := []int{4, 47, 58, 93}

	result := tm.pickCellsToLearnOn(100, 0, winnerCells, connections)
	sort.Ints(result)
	assert.Equal(t, []int{4, 47, 58, 93}, result)
	assert.Equal(t, []int{}, tm.pickCellsToLearnOn(0, 0, winnerCells, connections))
	assert.Equal(t, []int{4, 58}, tm.pickCellsToLearnOn(2, 0, winnerCells, connections))
}

func TestAdaptSegmentToMin(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections
	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.1)

	tm.adaptSegment(0, []int{}, connections)
	assert.Equal(t, 0.0, connections.DataForSynapse(0).Permanence)

	// // Now permanence should be at min
	tm.adaptSegment(0, []int{}, connections)
	assert.Equal(t, 0.0, connections.DataForSynapse(0).Permanence)

}

func TestAdaptSegmentToMax(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections
	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.9)

	tm.adaptSegment(0, []int{0}, connections)
	assert.Equal(t, 1.0, connections.DataForSynapse(0).Permanence)

	// Now permanence should be at max
	tm.adaptSegment(0, []int{0}, connections)
	assert.Equal(t, 1.0, connections.DataForSynapse(0).Permanence)

}

func TestLeastUsedCell(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tmp.ColumnDimensions = []int{2}
	tmp.CellsPerColumn = 2

	tm := NewTemporalMemory(tmp)

	connections := tm.Connections
	connections.CreateSegment(0)
	connections.CreateSynapse(0, 3, 0.3)

	for i := 0; i < 100; i++ {
		assert.Equal(t, 1, tm.getLeastUsedCell(0, connections))
	}

}

func TestAdaptSegment(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	tm.adaptSegment(0, []int{0, 1}, connections)

	assert.Equal(t, 0.7, connections.DataForSynapse(0).Permanence)
	assert.Equal(t, 0.5, connections.DataForSynapse(1).Permanence)
	assert.Equal(t, 0.8, connections.DataForSynapse(2).Permanence)

}

func TestGetConnectedActiveSynapsesForSegment(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(1)
	connections.CreateSynapse(1, 733, 0.7)
	connections.CreateSegment(8)
	connections.CreateSynapse(2, 486, 0.9)

	activeSynapsesForSegment := map[int][]int{
		0: {0, 1},
		1: {3},
	}

	assert.Equal(t, []int{0}, tm.getConnectedActiveSynapsesForSegment(0,
		activeSynapsesForSegment,
		0.5,
		connections))

	assert.Equal(t, []int{3}, tm.getConnectedActiveSynapsesForSegment(1,
		activeSynapsesForSegment,
		0.5,
		connections))

}

func TestComputeActiveSynapsesNoActivity(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(1)
	connections.CreateSynapse(1, 733, 0.7)
	connections.CreateSegment(8)
	connections.CreateSynapse(2, 486, 0.9)
	activeCells := []int{}
	assert.Equal(t, map[int][]int{}, tm.computeActiveSynapses(activeCells, connections))

}

func TestGetBestMatchingSegment(t *testing.T) {

	tmp := NewTemporalMemoryParams()
	tmp.MinThreshold = 1
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(0)
	connections.CreateSynapse(1, 49, 0.9)
	connections.CreateSynapse(1, 3, 0.8)
	connections.CreateSegment(1)
	connections.CreateSynapse(2, 733, 0.7)
	connections.CreateSegment(8)
	connections.CreateSynapse(3, 486, 0.9)

	activeSynapsesForSegment := map[int][]int{
		0: []int{0, 1},
		1: []int{3},
		2: []int{5},
	}

	bestCell, connectedSyns := tm.getBestMatchingSegment(0, activeSynapsesForSegment, connections)
	assert.Equal(t, 0, bestCell)
	assert.Equal(t, []int{0, 1}, connectedSyns)

	bestCell, connectedSyns = tm.getBestMatchingSegment(1, activeSynapsesForSegment, connections)
	assert.Equal(t, 2, bestCell)
	assert.Equal(t, []int{5}, connectedSyns)

	bestCell, connectedSyns = tm.getBestMatchingSegment(8, activeSynapsesForSegment, connections)
	assert.Equal(t, -1, bestCell)
	assert.Equal(t, []int(nil), connectedSyns)

	bestCell, connectedSyns = tm.getBestMatchingSegment(100, activeSynapsesForSegment, connections)
	assert.Equal(t, -1, bestCell)
	assert.Equal(t, []int(nil), connectedSyns)

}

func TestGetBestMatchingCellFewestSegments(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tmp.ColumnDimensions = []int{2}
	tmp.CellsPerColumn = 2
	tmp.MinThreshold = 1
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 3, 0.3)
	activeSynapsesForSegment := map[int][]int{}

	for i := 0; i < 100; i++ {
		// Never pick cell 0, always pick cell 1
		cell, _ := tm.getBestMatchingCell(0, activeSynapsesForSegment, connections)
		assert.Equal(t, 1, cell)
	}

}

func TestGetBestMatchingCell(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tmp.MinThreshold = 1
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(0)
	connections.CreateSynapse(1, 49, 0.9)
	connections.CreateSynapse(1, 3, 0.8)
	connections.CreateSegment(1)
	connections.CreateSynapse(2, 733, 0.7)
	connections.CreateSegment(108)
	connections.CreateSynapse(3, 486, 0.9)

	activeSynapsesForSegment := map[int][]int{
		0: []int{0, 1},
		1: []int{3},
		2: []int{5},
	}

	bestCell, bestSeg := tm.getBestMatchingCell(0, activeSynapsesForSegment, connections)
	assert.Equal(t, 0, bestCell)
	assert.Equal(t, 0, bestSeg)

	//randomly picked
	bestCell, bestSeg = tm.getBestMatchingCell(3, activeSynapsesForSegment, connections)
	assert.Equal(t, 98, bestCell)
	assert.Equal(t, -1, bestSeg)

	//randomly picked
	bestCell, bestSeg = tm.getBestMatchingCell(999, activeSynapsesForSegment, connections)
	assert.Equal(t, 31970, bestCell)
	assert.Equal(t, -1, bestSeg)

}

func TestComputeActiveSynapses(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	//tmp.MinThreshold = 1
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(1)
	connections.CreateSynapse(1, 733, 0.7)
	connections.CreateSegment(8)
	connections.CreateSynapse(2, 486, 0.9)
	activeCells := []int{23, 37, 733, 4973}

	expected := map[int][]int{
		0: []int{0, 1},
		1: []int{3},
	}
	assert.Equal(t, expected, tm.computeActiveSynapses(activeCells, connections))

}

func TestComputePredictiveCells(t *testing.T) {

	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections

	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.5)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(1)
	connections.CreateSynapse(1, 733, 0.7)
	connections.CreateSynapse(1, 733, 0.4)
	connections.CreateSegment(1)
	connections.CreateSynapse(2, 974, 0.9)
	connections.CreateSegment(8)
	connections.CreateSynapse(3, 486, 0.9)
	connections.CreateSegment(100)

	activeSynapsesForSegment := map[int][]int{
		0: []int{0, 1},
		1: []int{3, 4},
		2: []int{5},
	}

	activeSegments, predictiveCells := tm.computePredictiveCells(activeSynapsesForSegment, connections)
	//TODO: numentas returns [0]
	assert.Equal(t, []int(nil), activeSegments)
	assert.Equal(t, []int(nil), predictiveCells)

}

func TestLearnOnSegments(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections
	connections.CreateSegment(0)
	connections.CreateSynapse(0, 23, 0.6)
	connections.CreateSynapse(0, 37, 0.4)
	connections.CreateSynapse(0, 477, 0.9)
	connections.CreateSegment(1)
	connections.CreateSynapse(1, 733, 0.7)
	connections.CreateSegment(8)
	connections.CreateSynapse(2, 486, 0.9)
	connections.CreateSegment(100)

	prevActiveSegments := []int{0, 2}
	learningSegments := []int{1, 3}

	prevActiveSynapsesForSegment := map[int][]int{
		0: []int{0, 1},
		1: []int{3},
	}

	winnerCells := []int{0}

	prevWinnerCells := []int{10, 11, 12, 13, 14}

	tm.learnOnSegments(prevActiveSegments,
		learningSegments,
		prevActiveSynapsesForSegment,
		winnerCells,
		prevWinnerCells,
		connections)

	//Check segment 0
	assert.Equal(t, 0.7, connections.DataForSynapse(0).Permanence)
	assert.Equal(t, 0.5, connections.DataForSynapse(1).Permanence)
	assert.Equal(t, 0.8, connections.DataForSynapse(2).Permanence)

	//Check segment 1
	assert.InEpsilon(t, 0.8, connections.DataForSynapse(3).Permanence, 0.1)
	assert.Equal(t, 2, len(connections.synapsesForSegment[1]))

	//Check segment 2
	assert.Equal(t, 0.9, connections.DataForSynapse(4).Permanence)
	assert.Equal(t, 1, len(connections.synapsesForSegment[2]))

	// Check segment 3
	assert.Equal(t, 1, len(connections.synapsesForSegment[3]))

}
