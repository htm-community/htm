package htm

import (
	"testing"
)

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
