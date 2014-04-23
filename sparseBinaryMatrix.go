package htm

import (
//"math"
)

//Entries are positions of non-zero values
type SparseEntry struct {
	Row int
	Col int
}

//Sparse binary matrix stores indexes of non-zero entries in matrix
//to conserve space
type SparseBinaryMatrix struct {
	Width             int
	Height            int
	TotalNonZeroCount int
	Entries           []SparseEntry
}

func NewSparseBinaryMatrix(width int, height int) SparseBinaryMatrix {
	m := SparseBinaryMatrix{}
	m.Height = height
	m.Width = width
	//Intialize with 70% sparsity
	//m.Entries = make([]SparseEntry, int(math.Ceil(width*height*0.3)))
	return m
}

// func NewRandSparseBinaryMatrix() *SparseBinaryMatrix {
// }

// func (sm *SparseBinaryMatrix) Resize(width int, height int) {
// }

//Get value at col,row position
func (sm *SparseBinaryMatrix) Get(col int, row int) bool {
	for _, val := range sm.Entries {
		if val.Row == row && val.Col == col {
			return true
		}
	}
	return false
}

func (sm *SparseBinaryMatrix) delete(col int, row int) {
	for idx, val := range sm.Entries {
		if val.Row == row && val.Col == col {
			sm.Entries = append(sm.Entries[:idx], sm.Entries[idx+1:]...)
			break
		}
	}

}

//Set value at col,row position
func (sm *SparseBinaryMatrix) Set(col int, row int, value bool) {
	if !value {
		sm.delete(col, row)
		return
	}

	if sm.Get(col, row) {
		return
	}

	newEntry := SparseEntry{}
	newEntry.Col = col
	newEntry.Row = row
	sm.Entries = append(sm.Entries, newEntry)

}
