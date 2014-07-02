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
	"testing"
)

func TestTp(t *testing.T) {
	//2048, 32, 0.21, 0.5, 11, 20, 0.1, 0.1, 1.0, 0.0, 14, False, 5, 2,
	//         False, 1960, 0, False, '', 3, 10, 5, 0, 32, 128, 32, 'normal'
	tps := NewTemporalPoolerParams()
	tps.Verbosity = 10
	tps.NumberOfCols = 100
	tps.CellsPerColumn = 10
	//tps.ActivationThreshold = 2
	//tps.MinThreshold = 2
	tps.CollectStats = true
	tp := NewTemporalPooler(*tps)

	odd := GenerateRandSequence(100, 50)
	even := GenerateRandSequence(100, 50)

	for i := 0; i < 10; i++ {
		//input := make([]bool, 100)
		var input []bool
		// for j := 0; j < 25; j++ {
		// 	ind := rand.Intn(500)
		// 	input[ind] = true
		// }

		if i%2 == 0 {
			input = even
		} else {
			input = odd
		}

		// input[0] = true
		// input[1] = true
		// input[2] = true
		// input[3] = true
		// input[4] = true
		// input[5] = true
		// input[6] = true
		// input[7] = true
		// input[8] = true
		// input[9] = true
		// input[10] = true
		// input[11] = true
		// input[12] = true
		// input[13] = true
		// input[14] = true
		// input[15] = true
		// input[16] = true
		// input[17] = true
		// input[18] = true
		// input[19] = true
		// input[20] = true
		out := tp.Compute(input, true, true)
		fmt.Println("output", OnIndices(out))
	}

	p := tp.Predict(3)
	//s := p.SparseMatrix()

	for r := 0; r < p.Rows(); r++ {
		for c := 0; c < p.Cols(); c++ {
			if p.Get(r, c) != 0 {
				fmt.Println("predicted [%v,%v]", r, c)
			}
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
