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
