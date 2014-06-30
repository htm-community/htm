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
	tps := NewTemporalPoolerParams()
	tps.Verbosity = 6
	tp := NewTemportalPooler(*tps)

	for i := 0; i < 10; i++ {
		input := make([]bool, 500)

		for j := 0; j < 25; j++ {
			ind := rand.Intn(500)
			input[ind] = true
		}

		out := tp.compute(input, true, true)
		fmt.Println(Bool2Int(out))
	}

}

func GenerateSequence() {

}
