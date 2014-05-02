package htm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//Tests getting/setting values
func TestGetSet(t *testing.T) {

	sm := NewSparseBinaryMatrix(10, 10)
	sm.Set(2, 4, true)
	sm.Set(6, 5, true)
	sm.Set(7, 5, false)

	if !sm.Get(2, 4) {
		t.Errorf("Was false expected true @ [2,4]")
	}

	if !sm.Get(6, 5) {
		t.Errorf("Was false expected true @ [6,5]")
	}

	if sm.Get(7, 5) {
		t.Errorf("Was true expected false @ [7,5]")
	}

}

func TestRowReplace(t *testing.T) {
	sm := NewSparseBinaryMatrix(10, 10)
	sm.Set(2, 4, true)
	sm.Set(6, 5, true)
	sm.Set(7, 5, true)
	sm.Set(8, 8, true)

	if !sm.Get(8, 8) {
		t.Errorf("Was false expected true @ [8,8]")
	}

	newRow := make([]bool, 10)
	newRow[6] = true
	sm.ReplaceRow(8, newRow)

	if !sm.Get(8, 6) {
		t.Errorf("Was false expected true @ [8,6]")
	}

	if sm.Get(8, 8) {
		t.Errorf("Was true expected false @ [8,8]")
	}

}

func TestReplaceRowByIndices(t *testing.T) {
	sm := NewSparseBinaryMatrix(10, 10)

	indices := make([]int, 3)
	indices[0] = 3
	indices[1] = 9
	indices[2] = 6
	sm.ReplaceRowByIndices(4, indices)

	if !sm.Get(4, 3) {
		t.Errorf("Was false expected true @ [4,3]")
	}

	if !sm.Get(4, 9) {
		t.Errorf("Was false expected true @ [4,9]")
	}

	if !sm.Get(4, 6) {
		t.Errorf("Was false expected true @ [4,6]")
	}

	if sm.Get(4, 5) {
		t.Errorf("Was true expected false @ [4,5]")
	}

	if sm.Get(4, 0) {
		t.Errorf("Was true expected false @ [4,0]")
	}
}

func TestGetRowIndices(t *testing.T) {
	sm := NewSparseBinaryMatrix(10, 10)

	indices := make([]int, 3)
	indices[0] = 3
	indices[1] = 6
	indices[2] = 9
	sm.ReplaceRowByIndices(4, indices)

	indResult := sm.GetRowIndices(4)

	if len(indResult) != len(indices) {
		t.Errorf("Len was %v expected %v", len(indResult), len(indices))
	}

	for i := 0; i < 3; i++ {
		if indResult[i] != indices[i] {
			t.Errorf("Was %v expected %v", indResult, indices)
		}
	}

}

func TestGetRowAndSum(t *testing.T) {
	sm := NewSparseBinaryMatrix(4, 5)

	sm.SetRowFromDense(0, []bool{true, false, true, true, false})
	sm.SetRowFromDense(1, []bool{false, false, false, true, false})
	sm.SetRowFromDense(2, []bool{false, false, false, false, false})
	sm.SetRowFromDense(3, []bool{true, true, true, true, true})

	t.Log(sm.ToString())
	t.Log(sm.Entries)
	i := []bool{true, false, true, true, false}

	result := sm.RowAndSum(i)

	assert.Equal(t, 3, result[0])
	assert.Equal(t, 1, result[1])
	assert.Equal(t, 0, result[2])
	assert.Equal(t, 3, result[3])

}

func TestSetRowFromDense(t *testing.T) {

}

func TestNewFromDense(t *testing.T) {
	sbm := NewSparseBinaryMatrixFromDense([][]bool{
		{true, true, true},
		{false, false, false},
		{false, true, false},
		{true, false, true},
	})

	assert.Equal(t, 4, sbm.Height)
	assert.Equal(t, 3, sbm.Width)
	assert.Equal(t, true, sbm.Get(3, 2))
	assert.Equal(t, []bool{false, true, false}, sbm.GetDenseRow(2))

}
