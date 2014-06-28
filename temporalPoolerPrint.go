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

type SegmentStats struct {
	NumSegments        int
	NumSynapses        int
	NumActiveSynapses  int
	DistSegSizes       float64
	DistNumSegsPerCell float64
	DistPermValues     float64
	DistAges           float64
}

/*
Returns information about the distribution of segments, synapses and
permanence values in the current TP. If requested, also returns information
regarding the number of currently active segments and synapses.

*/

func (tp *TemporalPooler) calcSegmentStats(collectActiveData bool) SegmentStats {
	result := SegmentStats{}

	var distNSegsPerCell map[int]int
	var distSegSizes map[int]int
	var distPermValues map[int]int

	numAgeBuckets := 20
	ageBucketSize := int((self.lrnIterationIdx + 20) / 20)

	distAges := make(map[int]int, numAgeBuckets)
	distAgesLabels := make([]string, numAgeBuckets)
	for i := 0; i < numAgeBuckets; i++ {
		distAgesLabels[i] = fmt.Sprintf("%v-%v", i*ageBucketSize, (i+1)*ageBucketSize-1)
	}

	for ci, col := range tp.cells {
		for _, cell := range col {

			nSegmentsThisCell := len(cell)
			result.NumSegments += nSegmentsThisCell

			if val, ok := distNSegsPerCell[nSegmentsThisCell]; ok {
				distNSegsPerCell[nSegmentsThisCell]++
			} else {
				distNSegsPerCell[nSegmentsThisCell] = 1
			}

			for _, seg := range cell {
				nSynapsesThisSeg := len(seg.syns)
				nSynapses += nSynapsesThisSeg

				if val, ok := distSegSizes[nSynapsesThisSeg]; ok {
					distSegSizes[nSynapsesThisSeg]++
				} else {
					distSegSizes[nSynapsesThisSeg] = 1
				}

				// Accumulate permanence value histogram
				for _, syn := range seg.syns {
					p := int(syn.Permanence * 10)
					if val, ok := distPermValues[p]; ok {
						distPermValues[p]++
					} else {
						distPermValues[p] = 1
					}
				}

				// Accumulate segment age histogram
				age := tp.lrnIterationIdx - seg.lastActiveIteration
				ageBucket := int(age / ageBucketSize)
				distAges[ageBucket]++

				// Get active synapse statistics if requested
				if collectActiveData {
					if tp.isSegmentActive(seg, tp.DynamicState.infActiveState) {
						result.NumSegments++
					}
					for _, syn := range seg.syns {
						if tp.DynamicState.infActiveState.Get(syn.SrcCellIdx, syn.SrcCellCol) {
							result.NumActiveSynapses++
						}
					}
				}

			}
		}
	}

	return result
}

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
	counts := make([]int, tp.DynamicState.infActiveState.Height)
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

	stats := tp.calcSegmentStats(true)
	fmt.Println("numSegments", stats.numSegments)

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("----- infActiveState (%v on) ------", tp.DynamicState.infActiveState.TotalNonZeroCount())

}
