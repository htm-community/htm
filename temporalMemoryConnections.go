package htm

// import (
// 	"fmt"
// 	"github.com/cznic/mathutil"
// 	"github.com/zacg/floats"
// 	"github.com/zacg/go.matrix"
// 	"github.com/zacg/htm/utils"
// 	//"math"
// 	"math/rand"
// 	//"sort"
// )

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

func NewTemporalMemoryConnections(tmParams *TemporalMemoryParams) *TemporalMemoryConnections {
	c := new(TemporalMemoryConnections)
	c.maxSynapseCount = tmParams.MaxNewSynapseCount
	c.CellsPerColumn = tmParams.CellsPerColumn
	c.ColumnDimensions = tmParams.ColumnDimensions

	c.synapses = make([]*TmSynapse, 0, c.maxSynapseCount)

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
