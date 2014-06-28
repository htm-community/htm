//
// Code related to temporal pooler printing
//

package htm

import (
	"fmt"
	"github.com/cznic/mathutil"
	//"github.com/skelterjohn/go.matrix"
	//"math"
	//"math/rand"
	//"sort"
	//"github.com/gonum/floats"
	"github.com/zacg/ints"
)

/*
 Called at the end of inference to print out various diagnostic
information based on the current verbosity level.
*/

func (tp *TemporalPooler) printComputeEnd() {

	if tp.params.Verbosity < 3 {
		return
	}

	fmt.Println("----- computeEnd summary: ")
	fmt.Println("learn:", learn)
	bursting := 0
	counts := maeke([]int, tp.DynamicState.infActiveState.Height)
	for idx, val := range tp.DynamicState.infActiveState.Entries {
		counts[val.Row]++
		if counts[val.Row] == tp.DynamicState.infActiveState.Width {
			bursting++
		}
	}
	fmt.Println("numBurstingCols:", bursting)
	fmt.Println("curPredScore2:", tp.internalStats.CurPredictionScore2)
	fmt.Println("curFalsePosScore", tp.internalStats.CurFalsePositiveScore)
	fmt.Println("1-curFalseNegScore", 1-tp.internalStats.CurFalseNegativeScore)
	fmt.Println("numSegments:", tp.addToSegmentUpdates(c, i, segUpdate))
	fmt.Println("avgLearnedSeqLength", tp.avgLearnedSeqLength)

	//print "numSegments: ", self.getNumSegments(),

}
