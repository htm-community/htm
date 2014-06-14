package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/skelterjohn/go.matrix"
	"math"
	"math/rand"
	"sort"
)

type SegmentUpdate struct {
	columnIdx        int
	cellIdx          int
	segment          *Segment
	activeSynapses   []int
	sequenceSegment  bool
	phase1Flag       bool
	weaklyPredicting bool
	segmentUpdates   map[TupleInt][]UpdateState
	lrnIterationIdx  int
}

type UpdateState struct {
	IterationIdx int
	Update       *SegmentUpdate
}

/*
 Store a dated potential segment update. The "date" (iteration index) is used
later to determine whether the update is too old and should be forgotten.
This is controlled by parameter segUpdateValidDuration.
*/

func (su *SegmentUpdate) addToSegmentUpdates(c, i int, segUpdate *SegmentUpdate) {
	if segUpdate == nil || len(segUpdate.activeSynapses) == 0 {
		return
	}

	// key = (column index, cell index in column)
	key := TupleInt{c, i}

	newUpdate := UpdateState{su.lrnIterationIdx, segUpdate}
	if _, ok := su.segmentUpdates[key]; ok {
		su.segmentUpdates[key] = append(su.segmentUpdates[key], newUpdate)
	} else {
		su.segmentUpdates[key] = []UpdateState{newUpdate}
	}

}
