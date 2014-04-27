package htm

import (
	//"fmt"
	"github.com/skelterjohn/go.matrix"
	"math"
	"testing"
)

/*
def testInitPermanence1(self):
    """
test initial permanence generation. ensure that
a correct amount of synapses are initialized in
a connected state, with permanence values drawn from
the correct ranges
"""
    sp = self._sp
    sp._inputDimensions = numpy.array([10])
    sp._numInputs = 10
    sp._raisePermanenceToThreshold = Mock()

    sp._potentialRadius = 2
    connectedPct = 1
    mask = numpy.array([1, 1, 1, 0, 0, 0, 0, 0, 1, 1])
    perm = sp._initPermanence(mask, connectedPct)
    connected = (perm >= sp._synPermConnected).astype(int)
    numcon = (connected.nonzero()[0]).size
    self.assertEqual(numcon, 5)
    maxThresh = sp._synPermConnected + sp._synPermActiveInc/4
    self.assertEqual((perm <= maxThresh).all(), True)

    connectedPct = 0
    perm = sp._initPermanence(mask, connectedPct)
    connected = (perm >= sp._synPermConnected).astype(int)
    numcon = (connected.nonzero()[0]).size
    self.assertEqual(numcon, 0)

    connectedPct = 0.5
    sp._potentialRadius = 100
    sp._numInputs = 100
    mask = numpy.ones(100)
    perm = sp._initPermanence(mask, connectedPct)
    connected = (perm >= sp._synPermConnected).astype(int)
    numcon = (connected.nonzero()[0]).size
    self.assertGreater(numcon, 0)
    self.assertLess(numcon, sp._numInputs)

    minThresh = sp._synPermActiveInc / 2.0
    connThresh = sp._synPermConnected
    self.assertEqual(numpy.logical_and((perm >= minThresh),
                                       (perm < connThresh)).any(), True)
*/

// func TestPermanenceInit(t *testing.T){
// 	sp = self._sp
//     sp._inputDimensions = numpy.array([10])
//     sp._numInputs = 10
//     sp._raisePermanenceToThreshold = Mock()

// }

func TestRaisePermanenceThreshold(t *testing.T) {

	sp := SpatialPooler{}
	sp.InputDimensions = ITuple{1, 5}
	sp.ColumnDimensions = ITuple{1, 5}
	sp.numColumns = 5
	sp.numInputs = 5
	sp.SynPermConnected = 0.1
	sp.StimulusThreshold = 3
	sp.SynPermBelowStimulusInc = 0.01
	sp.SynPermMin = 0
	sp.SynPermMax = 1

	elms := make(map[int]float64, 25)
	sp.permanences = matrix.MakeSparseMatrix(elms, 5, 5)

	p := [][]float64{
		{0.0, 0.11, 0.095, 0.092, 0.01},
		{0.12, 0.15, 0.02, 0.12, 0.09},
		{0.51, 0.081, 0.025, 0.089, 0.31},
		{0.18, 0.0601, 0.11, 0.011, 0.03},
		{0.011, 0.011, 0.011, 0.011, 0.011},
	}
	AddDenseToSparseHelper(p, sp.permanences)

	fmt.Println("permanences vec", SparseMatrixToArray(sp.permanences.GetRowVector(2)))

	sp.connectedSynapses = NewSparseBinaryMatrixFromDense([][]bool{
		{false, true, false, false, false},
		{true, true, false, true, false},
		{true, false, false, false, true},
		{true, false, true, false, false},
		{false, false, false, false, false},
	})
	sp.connectedCounts = []int{1, 3, 2, 2, 0}

	truePermanences := [][]float64{
		{0.01, 0.12, 0.105, 0.102, 0.02},
		{0.12, 0.15, 0.02, 0.12, 0.09},
		{0.53, 0.101, 0.045, 0.109, 0.33},
		{0.22, 0.1001, 0.15, 0.051, 0.07},
		{0.101, 0.101, 0.101, 0.101, 0.101},
	}

	maskPP := []int{0, 1, 2, 3, 4}

	for i := 0; i < sp.numColumns; i++ {
		fmt.Println("calling:", i)
		perm := SparseMatrixToArray(sp.permanences.GetRowVector(i))
		sp.raisePermanenceToThreshold(perm, maskPP)
		for j := 0; j < sp.numInputs; j++ {
			//if truePermanences[i][j] != perm[j] {
			if !AlmostEqual(truePermanences[i][j], perm[j]) {
				t.Errorf("truePermances: %v != perm: %v", truePermanences[i][j], perm[j])
			}
		}
	}

}

func AlmostEqual(a, b float64) bool {
	ar := RoundPrec(a, 2)
	br := RoundPrec(b, 2)
	return ar == br
}

func SparseMatrixToArray(m *matrix.SparseMatrix) []float64 {
	result := make([]float64, m.Cols())
	for i := 0; i < m.Cols(); i++ {
		result[i] = m.Get(0, i)
	}
	return result
}

func AddDenseToSparseHelper(dense [][]float64, m *matrix.SparseMatrix) {
	for r := 0; r < len(dense); r++ {
		for c := 0; c < len(dense[r]); c++ {
			m.Set(r, c, dense[r][c])
		}
	}
}

func RoundPrec(x float64, prec int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow * sign
}
