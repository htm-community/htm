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

type Synapse struct {
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

	segments []Segment
	synapses []Synapse

	synapsesForSegment [][]int

	segmentIndex int
	synIndex     int
}

func (tmc *TemporalMemoryConnections) nextSegmentIndex() int {
	idx := tmc.segmentIndex
	tmc.segmentIndex++
	return idx
}

func (tmc *TemporalMemoryConnections) nextSynapseIndex() int {
	idx := tmc.synIndex
	tmc.synIndex++
	return idx
}

func (tmc *TemporalMemoryConnections) CreateSynapse(segment int, sourceCell int, permanence float64) {
	syn = tmc.nextSynapseIndex()
	tmc.synapses[syn] = Synapse{segment, sourceCell, permanence}

}
