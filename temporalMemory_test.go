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

func TestAdaptSegmentToMax(t *testing.T) {
	tmp := NewTemporalMemoryParams()
	tm := NewTemporalMemory(tmp)
	connections := tm.Connections
	connections = tm.Connections
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
