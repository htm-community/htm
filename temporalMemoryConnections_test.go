package htm

import (
	"github.com/zacg/testify/assert"
	"testing"
)

func TestNumColumns(t *testing.T) {
	c := NewTemporalMemoryConnections(0, 32, []int{64, 64})
	assert.Equal(t, 64*64, c.NumberOfColumns())
}

func TestNumCells(t *testing.T) {
	c := NewTemporalMemoryConnections(0, 32, []int{64, 64})
	assert.Equal(t, 64*64*32, c.NumberOfcells())
}

func TestUpdateSynapsePermanence(t *testing.T) {
	c := NewTemporalMemoryConnections(1000, 32, []int{64, 64})
	c.CreateSegment(0)
	c.CreateSynapse(0, 483, 0.1284)
	c.UpdateSynapsePermanence(0, 0.2496)
	assert.Equal(t, 0.2496, c.DataForSynapse(0).Permanence)
}

func TestCellsForColumn1D(t *testing.T) {
	c := NewTemporalMemoryConnections(1000, 5, []int{2048})
	expectedCells := []int{5, 6, 7, 8, 9}
	assert.Equal(t, expectedCells, c.CellsForColumn(1))
}

func TestCellsForColumn2D(t *testing.T) {
	c := NewTemporalMemoryConnections(1000, 4, []int{64, 64})
	expectedCells := []int{256, 257, 258, 259}
	assert.Equal(t, expectedCells, c.CellsForColumn(64))
}
