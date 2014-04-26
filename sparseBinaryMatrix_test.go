package htm

import (
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
