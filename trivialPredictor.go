package htm

import (
	//"fmt"
	"github.com/cznic/mathutil"
	//"github.com/skelterjohn/go.matrix"
	"math"
	//"math/rand"
	//"sort"
	"github.com/gonum/floats"
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

type TrivialPredictor struct {
	NumOfCols     int
	Methods       []PredictorMethod
	Verbosity     int
	InternalStats map[string]int
}
