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
