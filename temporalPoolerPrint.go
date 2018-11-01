//
// Code related to temporal pooler printing
//

package htm

import (
	"fmt"
	//"github.com/cznic/mathutil"
	//"github.com/zacg/go.matrix"
	//"math"
	//"math/rand"
	//"sort"
	//"github.com/gonum/floats"
	//"github.com/zacg/ints"
	"github.com/nupic-community/htm/utils"
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
	ageBucketSize := int((tp.lrnIterationIdx + 20) / 20)

	distAges := make(map[int]int, numAgeBuckets)
	distAgesLabels := make([]string, numAgeBuckets)
	for i := 0; i < numAgeBuckets; i++ {
		distAgesLabels[i] = fmt.Sprintf("%v-%v", i*ageBucketSize, (i+1)*ageBucketSize-1)
	}

	distNSegsPerCell = make(map[int]int, 1000)
	distSegSizes = make(map[int]int, 1000)
	distPermValues = make(map[int]int, 1000)

	for _, col := range tp.cells {
		for _, cell := range col {

			nSegmentsThisCell := len(cell)
			result.NumSegments += nSegmentsThisCell

			if _, ok := distNSegsPerCell[nSegmentsThisCell]; ok {
				distNSegsPerCell[nSegmentsThisCell]++
			} else {
				distNSegsPerCell[nSegmentsThisCell] = 1
			}

			for _, seg := range cell {
				nSynapsesThisSeg := len(seg.syns)
				result.NumSynapses += nSynapsesThisSeg

				if _, ok := distSegSizes[nSynapsesThisSeg]; ok {
					distSegSizes[nSynapsesThisSeg]++
				} else {
					distSegSizes[nSynapsesThisSeg] = 1
				}

				// Accumulate permanence value histogram
				for _, syn := range seg.syns {
					p := int(syn.Permanence * 10)
					if _, ok := distPermValues[p]; ok {
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
					if tp.isSegmentActive(seg, tp.DynamicState.InfActiveState) {
						result.NumSegments++
					}
					for _, syn := range seg.syns {
						if tp.DynamicState.InfActiveState.Get(syn.SrcCellIdx, syn.SrcCellCol) {
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
 Print the list of [column, cellIdx] indices for each of the active
cells in state.
*/
func (tp *TemporalPooler) printActiveIndices(state *SparseBinaryMatrix, andValues bool) {
	if state.TotalNonZeroCount() == 0 {
		fmt.Println("None")
		return
	}

	fmt.Println(state.Entries())

}

/*
	Prints a cels information
*/
func (tp *TemporalPooler) printCell(c int, i int, onlyActiveSegments bool) {

	cell := tp.cells[c][i]

	if len(cell) > 0 {
		fmt.Printf("Column: %v Cell: %v - %v segment(s)", c, i, len(cell))
		for idx, seg := range cell {
			isActive := tp.isSegmentActive(seg, tp.DynamicState.InfActiveState)
			if !onlyActiveSegments || isActive {
				str := " "
				if isActive {
					str = "*"
				}
				fmt.Printf("%vSeg: %v", str, idx)
				fmt.Println(seg.ToString())
			}
		}
	}

}

/*
 Print all cell information
*/
func (tp *TemporalPooler) printCells(predictedOnly bool) {

	if predictedOnly {
		fmt.Println("--- PREDICTED CELLS ---")
	} else {
		fmt.Println("--- ALL CELLS ---")
	}

	fmt.Println("Activation threshold:", tp.params.ActivationThreshold)
	fmt.Println("min threshold:", tp.params.MinThreshold)
	fmt.Println("connected perm:", tp.params.ConnectedPerm)

	for c, col := range tp.cells {
		for i := range col {
			if !predictedOnly || tp.DynamicState.InfPredictedState.Get(c, i) {
				tp.printCell(c, i, predictedOnly)
			}
		}
	}

}

/*
 Called at the end of inference to print out various diagnostic
information based on the current verbosity level.
*/
func (tp *TemporalPooler) printComputeEnd(output []bool, learn bool) {

	if tp.params.Verbosity < 3 {
		if tp.params.Verbosity >= 1 {
			fmt.Println("TP: learn:", learn)
			fmt.Printf("TP: active outputs(%v):\n", utils.CountTrue(output))
			fmt.Print(NewSparseBinaryMatrixFromDense1D(output,
				tp.params.NumberOfCols, tp.params.CellsPerColumn).ToString())
		}
		return
	}

	fmt.Println("----- computeEnd summary: ")
	fmt.Println("learn:", learn)
	bursting := 0
	counts := make([]int, tp.DynamicState.InfActiveState.Height)
	for _, val := range tp.DynamicState.InfActiveState.Entries() {
		counts[val.Row]++
		if counts[val.Row] == tp.DynamicState.InfActiveState.Width {
			bursting++
		}
	}
	fmt.Println("numBurstingCols:", bursting)
	fmt.Println("curPredScore2:", tp.internalStats.CurPredictionScore2)
	fmt.Println("curFalsePosScore", tp.internalStats.CurFalsePositiveScore)
	fmt.Println("1-curFalseNegScore", 1-tp.internalStats.CurFalseNegativeScore)
	fmt.Println("avgLearnedSeqLength", tp.avgLearnedSeqLength)

	stats := tp.calcSegmentStats(true)
	fmt.Println("numSegments", stats.NumSegments)

	fmt.Printf("----- InfActiveState (%v on) ------\n", tp.DynamicState.InfActiveState.TotalNonZeroCount())
	tp.printActiveIndices(tp.DynamicState.InfActiveState, false)

	if tp.params.Verbosity >= 6 {
		//tp.printState(tp.InfActiveState['t'])
		//fmt.Println(tp.DynamicState.InfActiveState.ToString())
	}

	fmt.Printf("----- InfPredictedState (%v on)-----\n", tp.DynamicState.InfPredictedState.TotalNonZeroCount())
	tp.printActiveIndices(tp.DynamicState.InfPredictedState, false)
	if tp.params.Verbosity >= 6 {
		//fmt.Println(tp.DynamicState.InfPredictedState.ToString())
	}

	fmt.Printf("----- LrnActiveState (%v on) ------\n", tp.DynamicState.LrnActiveState.TotalNonZeroCount())
	tp.printActiveIndices(tp.DynamicState.LrnActiveState, false)
	if tp.params.Verbosity >= 6 {
		//fmt.Println(tp.DynamicState.LrnActiveState.ToString())
	}

	fmt.Printf("----- LrnPredictedState (%v on)-----\n", tp.DynamicState.LrnPredictedState.TotalNonZeroCount())
	tp.printActiveIndices(tp.DynamicState.LrnPredictedState, false)
	if tp.params.Verbosity >= 6 {
		//fmt.Println(tp.DynamicState.LrnPredictedState.ToString())
	}

	fmt.Println("----- CellConfidence -----")
	//tp.printActiveIndices(tp.DynamicState.CellConfidence, true)

	if tp.params.Verbosity >= 6 {
		//TODO: this
		//tp.printConfidence(tp.DynamicState.CellConfidence)
		for r := 0; r < tp.DynamicState.CellConfidence.Rows(); r++ {
			for c := 0; c < tp.DynamicState.CellConfidenceLast.Cols(); c++ {
				if tp.DynamicState.CellConfidence.Get(r, c) != 0 {
					fmt.Printf("[%v,%v,%v]", r, c, tp.DynamicState.CellConfidence.Get(r, c))
				}
			}
		}

	}

	fmt.Println("----- ColConfidence -----")
	//tp.printActiveIndices(tp.DynamicState.ColConfidence, true)
	fmt.Println("----- CellConfidence[t-1] for currently active cells -----")
	//cc := matrix.ZerosSparse(tp.DynamicState.CellConfidence.Rows(), tp.DynamicState.CellConfidence.Cols())
	for _, val := range tp.DynamicState.InfActiveState.Entries() {
		//cc.Set(val.Row, val.Col, tp.DynamicState.CellConfidence.Get(val.Row, val.Col))
		fmt.Printf("[%v,%v,%v]", val.Row, val.Col, tp.DynamicState.CellConfidence.Get(val.Row, val.Col))

	}
	//fmt.Println(cc.String())

	if tp.params.Verbosity == 4 {
		fmt.Println("Cells, predicted segments only:")
		tp.printCells(true)
	} else if tp.params.Verbosity >= 5 {
		fmt.Println("Cells, all segments:")
		tp.printCells(true)
	}

}
