package htm

import (
	// 	"fmt"
	// 	"github.com/cznic/mathutil"
	// 	"github.com/zacg/floats"
	// 	"github.com/zacg/go.matrix"
	//"github.com/zacg/htm/utils"
	"github.com/zacg/ints"
	// 	//"math"
	// 	"math/rand"
	// 	//"sort"
)

type TmSynapse struct {
	Segment    int
	SourceCell int
	Permanence float64
}

/*
 Structure holds data representing the connectivity of a layer of cells,
that the TM operates on.
*/
type TemporalMemoryConnections struct {
	ColumnDimensions []int
	CellsPerColumn   int

	segments []*Segment
	synapses []*TmSynapse

	synapsesForSegment    [][]int
	synapsesForSourceCell [][]int

	segmentsForCell [][]int

	segmentIndex int
	synIndex     int

	maxSynapseCount int
}

//Create a new temporal memory
func NewTemporalMemoryConnections(maxSynCount int, cellsPerColumn int, colDimensions []int) *TemporalMemoryConnections {
	if len(colDimensions) < 1 {
		panic("Column dimensions must be greater than 0")
	}

	if cellsPerColumn < 1 {
		panic("Number of cells per column must be greater than 0")
	}

	c := new(TemporalMemoryConnections)
	c.maxSynapseCount = maxSynCount
	c.CellsPerColumn = cellsPerColumn
	c.ColumnDimensions = colDimensions

	c.synapses = make([]*TmSynapse, 0, c.maxSynapseCount)
	//TODO: init segments

	return c
}

// func (tmc *TemporalMemoryConnections) nextSegmentIndex() int {
// 	idx := tmc.segmentIndex
// 	tmc.segmentIndex++
// 	return idx
// }

// func (tmc *TemporalMemoryConnections) nextSynapseIndex() int {
// 	idx := tmc.synIndex
// 	tmc.synIndex++
// 	return idx
// }

func (tmc *TemporalMemoryConnections) CreateSynapse(segment int, sourceCell int, permanence float64) *TmSynapse {
	syn := len(tmc.synapses)
	data := new(TmSynapse)
	data.Segment = segment
	data.SourceCell = sourceCell
	data.Permanence = permanence
	tmc.synapses = append(tmc.synapses, data)

	//Update indexes
	tmc.synapsesForSegment[segment] = append(tmc.synapsesForSegment[segment], syn)
	tmc.synapsesForSourceCell[sourceCell] = append(tmc.synapsesForSourceCell[sourceCell], syn)

	return data
}

//Creates a new segment on specified cell, returns segment index
func (tmc *TemporalMemoryConnections) CreateSegment(cell int) int {
	idx := len(tmc.segments)
	// Add data
	tmc.segments = append(tmc.segments, cell)
	tmc.segmentsForCell[cell] = append(tmc.segmentsForCell[idx], idx)
	return idx
}

//Updates the permanence for a synapse.
func (tmc *TemporalMemoryConnections) UpdateSynapsePermanence(synapse int, permanence float64) {
	tmc.validatePermanence(permanence)
	tmc.synapses[synapse].Permanence = permanence
}

//Returns the index of the column that a cell belongs to.
func (tmc *TemporalMemoryConnections) ColumnForCell(cell int) int {
	return int(cell / tmc.cellsPerColumn)
}

//Returns the indices of cells that belong to a column.
func (tmc *TemporalMemoryConnections) CellsForColumn(column int) []int {
	start := tmc.CellsPerColumn * column
	result := make([]int, tmc.CellsPerColumn)
	for idx, val := range result {
		result[idx] = start + idx
	}
	return result
}

//Returns the cell that a segment belongs to.
func (tmc *TemporalMemoryConnections) CellForSegment(segment int) int {
	return tmc.segments[segment]
}

//Returns the segments that belong to a cell.
func (tmc *TemporalMemoryConnections) SegmentsForCell(cell int) []int {
	return tmc.segmentsForCell[cell]
}

//Returns synapse data for specified index
func (tmc *TemporalMemoryConnections) DataForSynapse(synapse int) TmSynapse {
	return tmc.synapses[synapse]
}

//Returns the synapses on a segment.
func (tmc *TemporalMemoryConnections) SynapsesForSegment(segment int) []int {
	return tmc.synapsesForSegment[segment]
}

//Returns the synapses for the source cell that they synapse on.
func (tmc *TemporalMemoryConnections) SynapsesForSourceCell(sourceCell int) []int {
	return tmc.synapsesForSourceCell[sourceCell]
}

// Helpers

//Returns the number of columns in this layer.
func (tmc *TemporalMemoryConnections) NumberOfColumns() int {
	return ints.Prod(tmc.ColumnDimensions)
}

//Returns the number of cells in this layer.
func (tmc *TemporalMemoryConnections) NumberOfcells() int {
	return tmc.NumberOfColumns() * tmc.CellsPerColumn
}

//Validation

func (tmc *TemporalMemoryConnections) validatePermanence(permanence float64) {
	if permanence < 0 || permanence > 1 {
		panic("invalid permanence value")
	}
}
