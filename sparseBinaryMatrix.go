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

//Replaces specified row with values, assumes values is ordered
//correctly
func (sm *SparseBinaryMatrix) ReplaceRow(row int, values []bool) {
	sm.validateRowCol(row, len(values))

	for i := 0; i < sm.Width; i++ {
		sm.Set(i, row, values[i])
	}
}

//Returns dense row
func (sm *SparseBinaryMatrix) GetDenseRow(row int) []bool {
	sm.validateRow(row)
	result := make([]bool, sm.Width)

	for i := 0; i < len(sm.Entries); i++ {
		if sm.Entries[i].Row == row {
			result[sm.Entries[i].Col] = true
		}
	}

	return result
}

//Sets a sparse row from dense representation
func (sm *SparseBinaryMatrix) SetRowFromDense(row int, denseRow []bool) {
	sm.validateRowCol(row, len(denseRow))
	for i := 0; i < sm.Width; i++ {
		sm.Set(i, row, denseRow[i])
	}
}

func (sm *SparseBinaryMatrix) validateCol(col int) {
	if col > sm.Width {
		panic("Specified row is wider than matrix.")
	}
}

func (sm *SparseBinaryMatrix) validateRow(row int) {
	if row > sm.Height {
		panic("Specified row is out of bounds.")
	}
}

func (sm *SparseBinaryMatrix) validateRowCol(row int, col int) {
	sm.validateCol(col)
	sm.validateRow(row)
}
