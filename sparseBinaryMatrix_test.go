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
	sm.ReplaceRow(8, values)

	if !sm.Get(8, 6) {
		t.Errorf("Was false expected true @ [8,6]")
	}

	if sm.Get(8, 8) {
		t.Errorf("Was true expected false @ [8,8]")
	}

}
