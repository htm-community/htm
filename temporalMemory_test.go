package htm

import (
	//"fmt"
	"github.com/zacg/testify/assert"
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
