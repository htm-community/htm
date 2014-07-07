package htm

import (
	"fmt"
	//"github.com/cznic/mathutil"
	//"github.com/zacg/go.matrix"
	//"math"
	"math/rand"
	//"sort"
	//"github.com/gonum/floats"
	//"github.com/zacg/ints"
	"github.com/zacg/testify/assert"
	"testing"
)

func boolRange(start int, end int, length int) []bool {
	result := make([]bool, length)
	for i := start; i <= end; i++ {
		result[i] = true
	}

	return result
}

func TestLearnPredict(t *testing.T) {
	tps := NewTemporalPoolerParams()
	tps.Verbosity = 10
	tps.NumberOfCols = 50
	tps.CellsPerColumn = 2
	tps.ActivationThreshold = 8
	tps.MinThreshold = 10
	tps.InitialPerm = 0.5
	tps.ConnectedPerm = 0.5
	tps.NewSynapseCount = 10
	tps.PermanenceDec = 0.0
	tps.PermanenceInc = 0.1
	tps.GlobalDecay = 0
	tps.BurnIn = 1
	tps.PamLength = 10
	//tps.DoPooling = true

	tps.CollectStats = true
	tp := NewTemporalPooler(*tps)

	inputs := make([][]bool, 5)

	// inputs[0] = GenerateRandSequence(80, 50)
	// inputs[1] = GenerateRandSequence(80, 50)
	// inputs[2] = GenerateRandSequence(80, 50)
	inputs[0] = boolRange(0, 9, 50)
	inputs[1] = boolRange(10, 19, 50)
	inputs[2] = boolRange(20, 29, 50)
	inputs[3] = boolRange(30, 39, 50)
	inputs[4] = boolRange(40, 49, 50)

	//Learn 5 sequences above
	for i := 0; i < 10; i++ {
		for p := 0; p < 5; p++ {
			tp.Compute(inputs[p], true, false)
		}

		tp.Reset()
	}

	//Predict sequences
	for i := 0; i < 4; i++ {
		tp.Compute(inputs[i], false, true)
		p := tp.DynamicState.infPredictedState.Entries
		fmt.Println(p)
		assert.Equal(t, 10, len(p))
		for _, val := range p {
			next := i + 1
			if next > 4 {
				next = 4
			}
			assert.True(t, inputs[next][val.Row])
		}
	}

}

func GenerateRandSequence(size int, width int) []bool {
	input := make([]bool, size)
	for i := 0; i < width; i++ {
		ind := rand.Intn(size)
		input[ind] = true
	}

	return input
}
