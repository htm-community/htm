package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/skelterjohn/go.matrix"
	"math"
	"math/rand"
	"sort"
)

type Synapse struct {
	SrcCellCol int
	SrcCellIdx int
	Permanence float64
}

// The Segment struct is a container for all of the segment variables and
//the synapses it owns.
type Segment struct {
	tp                        *TemporalPooler
	segId                     int
	isSequenceSeg             bool
	lastActiveIteration       int
	positiveActivations       int
	totalActivations          int
	lastPosDutyCycle          int
	lastPosDutyCycleIteration int
	syns                      []Synapse
}

//Creates a new segment
func NewSegment(tp *TemporalPooler, isSequenceSeg bool) *Segment {
	seg := Segment{}
	seg.tp = tp
	seg.segId = tp.segID
	seg.isSequenceSeg = isSequenceSeg
	seg.lastActiveIteration = tp.lrnIterationIdx
	seg.positiveActivations = 1
	seg.totalActivations = 1

	seg.lastPosDutyCycle = 1.0 / tp.lrnIterationIdx
	seg.lastPosDutyCycleIteration = tp.lrnIterationIdx

	//TODO: initialize synapse collection

	return &seg
}
