package htm

import (
	//"fmt"
	"github.com/skelterjohn/go.matrix"
	"math"
	"testing"
)

func getConnected(perm []float64, sp *SpatialPooler) (int, []bool) {
	numcon := 0
	connected := make([]bool, len(perm))
	for i := 0; i < len(perm); i++ {
		if perm[i] >= sp.SynPermConnected {
			numcon++
			connected[i] = true
		} else {
			connected[i] = false
		}
	}

	return numcon, connected
}

func TestPermanenceInit(t *testing.T) {
	sp := SpatialPooler{}
	sp.InputDimensions = ITuple{1, 10}
	sp.numInputs = 10
	sp.SynPermConnected = 0.1
	sp.SynPermActiveInc = 0.1
	//sp.raisePermanenceToThreshold = Mock()

	sp.PotentialRadius = 2
	connectedPct := 1.0
	mask := []bool{true, true, true, false, false, false, false, false, true, true}
	perm := sp.initPermanence(mask, connectedPct)
	numcon, connected := getConnected(perm, &sp)

	if numcon != 5 {
		t.Errorf("numcon was %v expected 5", numcon)
	}
	maxThresh := sp.SynPermConnected + sp.SynPermActiveInc/4

	for i := 0; i < len(perm); i++ {
		if perm[i] > maxThresh {
			t.Errorf("perm %v was %v higher than threshold", i, perm[i])
		}
	}

	connectedPct = 0
	numcon = 0
	perm = sp.initPermanence(mask, connectedPct)
	numcon, connected = getConnected(perm, &sp)
	if numcon != 0 {
		t.Errorf("numcon was %v expected 0", numcon)
	}

	if len(connected) != 5 {
		return
	}

	connectedPct = 0.5
	sp.PotentialRadius = 100
	sp.numInputs = 100
	mask = make([]bool, 100)
	for i := 0; i < len(mask); i++ {
		mask[i] = true
	}

	perm = sp.initPermanence(mask, connectedPct)
	numcon, connected = getConnected(perm, &sp)

	if !(numcon > 0) {
		t.Errorf("numcon was %v expected greater than 0", numcon)
	}

	if numcon >= sp.numInputs {
		t.Errorf("numcon was %v expected less than inputs count", numcon)
	}

	minThresh := sp.SynPermActiveInc / 2.0
	connThresh := sp.SynPermConnected

	for i := 0; i < len(perm); i++ {
		if perm[i] < minThresh {
			t.Errorf("perm %v was %v less than min threshold", i, perm[i])
		}
		if perm[i] >= connThresh {
			t.Errorf("perm %v was %v not less than connection threshold", i, perm[i])
		}
	}

}

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

func TestStripNever(t *testing.T) {
	sp := SpatialPooler{}

	sp.activeDutyCycles = []float64{0.5, 0.1, 0, 0.2, 0.4, 0}
	activeColumns := []int{0, 1, 2, 4}
	stripped := sp.stripNeverLearned(activeColumns)
	trueStripped := []int{0, 1, 4}
	t.Logf("stripped", stripped)
	for i := 0; i < len(trueStripped); i++ {
		if stripped[i] != trueStripped[i] {
			t.Errorf("stripped %v was %v expected %v", i, stripped[i], trueStripped[i])
		}
	}

	sp.activeDutyCycles = []float64{0.9, 0, 0, 0, 0.4, 0.3}
	activeColumns = []int{0, 1, 2, 3, 4, 5}
	stripped = sp.stripNeverLearned(activeColumns)
	trueStripped = []int{0, 4, 5}
	for i := 0; i < len(trueStripped); i++ {
		if stripped[i] != trueStripped[i] {
			t.Errorf("stripped %v was %v expected %v", i, stripped[i], trueStripped[i])
		}
	}

	sp.activeDutyCycles = []float64{0, 0, 0, 0, 0, 0}
	activeColumns = []int{0, 1, 2, 3, 4, 5}
	stripped = sp.stripNeverLearned(activeColumns)
	if len(stripped) != 0 {
		t.Errorf("Expected empty stripped was %v", stripped)
	}

	sp.activeDutyCycles = []float64{1, 1, 1, 1, 1, 1}
	activeColumns = []int{0, 1, 2, 3, 4, 5}
	stripped = sp.stripNeverLearned(activeColumns)
	trueStripped = []int{0, 1, 2, 3, 4, 5}
	for i := 0; i < len(trueStripped); i++ {
		if stripped[i] != trueStripped[i] {
			t.Errorf("stripped %v was %v expected %v", i, stripped[i], trueStripped[i])
		}
	}

}

//----- Helper functions -------------

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
