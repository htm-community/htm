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
(n = half the number of average input columns on)
"random" - predict n random columns
"zeroth" - predict the n most common columns learned from the input
"last" - predict the last input
"all" - predict all columns
"lots" - predict the 2n most common columns learned from the input

Both "random" and "all" should give a prediction score of zero"
*/

type PredictorMethod int

const (
	Random PredictorMethod = 1
	Zeroth PredictorMethod = 2
	Last   PredictorMethod = 3
	All    PredictorMethod = 4
	Lots   PredictorMethod = 5
)

type TrivialPredictorState struct {
	ActiveState        []bool
	ActiveStateLast    []bool
	PredictedState     []bool
	PredictedStateLast []bool
	Confidence         []float64
	ConfidenceLast     []float64
}

type TrivialPredictor struct {
	NumOfCols      int
	Methods        []PredictorMethod
	Verbosity      int
	InternalStats  map[PredictorMethod]*TpStats
	State          map[PredictorMethod]TrivialPredictorState
	ColumnCount    []int
	AverageDensity float64
}

func MakeTrivialPredictor(numberOfCols int, methods []PredictorMethod) *TrivialPredictor {
	tp := new(TrivialPredictor)

	for _, method := range methods {
		tps := TrivialPredictorState{}
		tps.ActiveState = make([]bool, numberOfCols)
		tps.ActiveStateLast = make([]bool, numberOfCols)
		tps.Confidence = make([]float64, numberOfCols)
		tps.ConfidenceLast = make([]float64, numberOfCols)
		tps.PredictedState = make([]bool, numberOfCols)
		tps.PredictedStateLast = make([]bool, numberOfCols)
		tp.State[method] = tps

		tp.InternalStats[method] = new(TpStats)
	}

	// Number of times each column has been active during learning
	tp.ColumnCount = make([]int, numberOfCols)

	// Running average of input density
	tp.AverageDensity = 0.05

	return tp
}

/*

*/

func (tp *TrivialPredictor) infer(activeColumns []int) {

	numColsToPredict := int(0.5 + tp.AverageDensity*float64(tp.NumOfCols))

	//for method in self.methods:
	for _, method := range tp.Methods {
		// Copy t-1 into t
		copy(tp.State[method].ActiveStateLast, tp.State[method].ActiveState)
		copy(tp.State[method].PredictedStateLast, tp.State[method].PredictedState)
		copy(tp.State[method].ConfidenceLast, tp.State[method].Confidence)

		FillSliceBool(tp.State[method].ActiveState, false)
		FillSliceBool(tp.State[method].PredictedState, false)
		FillSliceFloat64(tp.State[method].Confidence, 0.0)

		for _, val := range activeColumns {
			tp.State[method].ActiveState[val] = true
		}

		var predictedCols []int

		switch method {
		case Random:
			// Randomly predict N columns
			//predictedCols = RandomInts(numColsToPredict, tp.NumOfCols)
			break
		case Zeroth:
			// Always predict the top N most frequent columns
			var inds []int
			ints.Argsort(tp.ColumnCount, inds)
			predictedCols = inds[len(inds)-numColsToPredict:]
			break
		case Last:
			// Always predict the last input
			for idx, val := range tp.State[method].ActiveState {
				if val {
					predictedCols = append(predictedCols, idx)
				}
			}
			break
		case All:
			// Always predict all columns
			for i := 0; i < tp.NumOfCols; i++ {
				predictedCols = append(predictedCols, i)
			}
			break
		case Lots:
			// Always predict 2 * the top N most frequent columns
			numColsToPredict := mathutil.Min(2*numColsToPredict, tp.NumOfCols)
			var inds []int
			ints.Argsort(tp.ColumnCount, inds)
			predictedCols = inds[len(inds)-numColsToPredict:]

			break
		default:
			panic("prediction method not implemented")
		}

		for _, val := range predictedCols {
			tp.State[method].PredictedState[val] = true
			tp.State[method].Confidence[val] = 1.0
		}

		if tp.Verbosity > 1 {
			fmt.Println("Random prediction:", method)
			fmt.Println(" numColsToPredict:", numColsToPredict)
			fmt.Println(predictedCols)
		}

	}

}

/*
 Do one iteration of the temporal pooler learning.
Returns TP output
*/

func (tp *TrivialPredictor) learn(activeColumns []int) {
	// Running average of bottom up density
	density := float64(len(activeColumns)) / float64(tp.NumOfCols)

	tp.AverageDensity = 0.95*tp.AverageDensity + 0.05*density

	// Running count of how often each column has been active
	for _, val := range activeColumns {
		tp.ColumnCount[val]++
	}

	// Do "inference"
	tp.infer(activeColumns)
}

/*
Reset the state of all cells.
This is normally used between sequences while training. All internal states
are reset to 0.
*/

func (tp *TrivialPredictor) reset() {

	for _, method := range tp.Methods {

		FillSliceBool(tp.State[method].ActiveState, false)
		FillSliceBool(tp.State[method].ActiveStateLast, false)
		FillSliceBool(tp.State[method].PredictedState, false)
		FillSliceBool(tp.State[method].PredictedStateLast, false)
		FillSliceFloat64(tp.State[method].Confidence, 0.0)
		FillSliceFloat64(tp.State[method].ConfidenceLast, 0.0)

		stats := tp.InternalStats[method]
		stats.NInfersSinceReset = 0
		stats.CurPredictionScore = 0.0
		stats.CurPredictionScore2 = 0.0
		stats.FalseNegativeScoreTotal = 0.0
		stats.FalsePositiveScoreTotal = 0.0
		stats.CurExtra = 0.0
		stats.CurMissing = 0.0
		tp.InternalStats[method] = stats
	}

}

/*
Reset the learning and inference stats. This will usually be called by
user code at the start of each inference run (for a particular data set).
*/

func (tp *TrivialPredictor) resetStats() {

	tp.reset()

	//Additionally, reset all of the "total" values
	for _, method := range tp.Methods {

		stats := tp.InternalStats[method]
		stats.NInfersSinceReset = 0
		stats.NPredictions = 0
		stats.PredictionScoreTotal = 0
		stats.PredictionScoreTotal2 = 0
		stats.FalseNegativeScoreTotal = 0
		stats.FalsePositiveScoreTotal = 0
		stats.PctExtraTotal = 0.0
		stats.PctMissingTotal = 0.0
		stats.TotalMissing = 0.0
		stats.TotalExtra = 0.0
		tp.InternalStats[method] = stats
	}
}
