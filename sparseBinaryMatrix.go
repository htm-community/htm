package htm

import (
	//"math"
	"bytes"
	"github.com/zacg/htm/utils"
)

//entries are positions of non-zero values
type SparseEntry struct {
	Row int
	Col int
}

//Sparse binary matrix stores indexes of non-zero entries in matrix
//to conserve space
type SparseBinaryMatrix struct {
	Width   int
	Height  int
	entries []SparseEntry
}

//Create new sparse binary matrix of specified size
func NewSparseBinaryMatrix(height, width int) *SparseBinaryMatrix {
	m := &SparseBinaryMatrix{}
	m.Height = height
	m.Width = width
	//Intialize with 70% sparsity
	//m.entries = make([]SparseEntry, int(math.Ceil(width*height*0.3)))
	return m
}

//Create sparse binary matrix from specified dense matrix
func NewSparseBinaryMatrixFromDense(values [][]bool) *SparseBinaryMatrix {
	if len(values) < 1 {
		panic("No values specified.")
	}
	m := &SparseBinaryMatrix{}
	m.Height = len(values)
	m.Width = len(values[0])

	for r := 0; r < m.Height; r++ {
		m.SetRowFromDense(r, values[r])
	}

	return m
}

//Create sparse binary matrix from specified dense matrix
func NewSparseBinaryMatrixFromDense1D(values []bool, rows, cols int) *SparseBinaryMatrix {
	if len(values) < 1 {
		panic("No values specified.")
	}
	if len(values) != rows*cols {
		panic("Invalid size")
	}

	m := new(SparseBinaryMatrix)
	m.Height = rows
	m.Width = cols

	for r := 0; r < m.Height; r++ {
		m.SetRowFromDense(r, values[r*cols:(r*cols)+cols])
	}

	return m
}

// Creates a sparse binary matrix from specified integer array
// (any values greater than 0 are true)
func NewSparseBinaryMatrixFromInts(values [][]int) *SparseBinaryMatrix {
	if len(values) < 1 {
		panic("No values specified.")
	}

	m := &SparseBinaryMatrix{}
	m.Height = len(values)
	m.Width = len(values[0])

	for r := 0; r < m.Height; r++ {
		for c := 0; c < m.Width; c++ {
			if values[r][c] > 0 {
				m.Set(r, c, true)
			}
		}
	}

	return m
}

// func NewRandSparseBinaryMatrix() *SparseBinaryMatrix {
// }

// func (sm *SparseBinaryMatrix) Resize(width int, height int) {
// }

//Returns all true/on indices
func (sm *SparseBinaryMatrix) Entries() []SparseEntry {
	return sm.entries
}

//Returns flattend dense represenation
func (sm *SparseBinaryMatrix) Flatten() []bool {
	result := make([]bool, sm.Height*sm.Width)
	for _, val := range sm.entries {
		result[(val.Row*sm.Width)+val.Col] = true
	}
	return result
}

//Get value at col,row position
func (sm *SparseBinaryMatrix) Get(row int, col int) bool {
	for _, val := range sm.entries {
		if val.Row == row && val.Col == col {
			return true
		}
	}
	return false
}

func (sm *SparseBinaryMatrix) delete(row int, col int) {
	for idx, val := range sm.entries {
		if val.Row == row && val.Col == col {
			sm.entries = append(sm.entries[:idx], sm.entries[idx+1:]...)
			break
		}
	}
}

//Set value at row,col position
func (sm *SparseBinaryMatrix) Set(row int, col int, value bool) {
	if !value {
		sm.delete(row, col)
		return
	}

	if sm.Get(row, col) {
		return
	}

	newEntry := SparseEntry{}
	newEntry.Col = col
	newEntry.Row = row
	sm.entries = append(sm.entries, newEntry)

}

//Replaces specified row with values, assumes values is ordered
//correctly
func (sm *SparseBinaryMatrix) ReplaceRow(row int, values []bool) {
	sm.validateRowCol(row, len(values))

	for i := 0; i < sm.Width; i++ {
		sm.Set(row, i, values[i])
	}
}

//Replaces row with true values at specified indices
func (sm *SparseBinaryMatrix) ReplaceRowByIndices(row int, indices []int) {
	sm.validateRow(row)

	for i := 0; i < sm.Width; i++ {
		val := false
		for x := 0; x < len(indices); x++ {
			if i == indices[x] {
				val = true
				break
			}
		}
		sm.Set(row, i, val)
	}
}

//Returns dense row
func (sm *SparseBinaryMatrix) GetDenseRow(row int) []bool {
	sm.validateRow(row)
	result := make([]bool, sm.Width)

	for i := 0; i < len(sm.entries); i++ {
		if sm.entries[i].Row == row {
			result[sm.entries[i].Col] = true
		}
	}

	return result
}

//Returns a rows "on" indices
func (sm *SparseBinaryMatrix) GetRowIndices(row int) []int {
	result := []int{}
	for i := 0; i < len(sm.entries); i++ {
		if sm.entries[i].Row == row {
			result = append(result, sm.entries[i].Col)
		}
	}
	return result
}

//Sets a sparse row from dense representation
func (sm *SparseBinaryMatrix) SetRowFromDense(row int, denseRow []bool) {
	sm.validateRowCol(row, len(denseRow))
	for i := 0; i < sm.Width; i++ {
		sm.Set(row, i, denseRow[i])
	}
}

//In a normal matrix this would be multiplication in binary terms
//we just and then sum the true entries
func (sm *SparseBinaryMatrix) RowAndSum(row []bool) []int {
	sm.validateCol(len(row))
	result := make([]int, sm.Height)

	for _, val := range sm.entries {
		if row[val.Col] {
			result[val.Row]++
		}
	}

	return result
}

//Returns row indexes with at least 1 true column
func (sm *SparseBinaryMatrix) NonZeroRows() []int {
	var result []int

	for _, val := range sm.entries {
		if !utils.ContainsInt(val.Row, result) {
			result = append(result, val.Row)
		}
	}

	return result
}

//Returns # of rows with at least 1 true value
func (sm *SparseBinaryMatrix) TotalTrueRows() int {
	var hitRows []int
	for _, val := range sm.entries {
		if !utils.ContainsInt(val.Row, hitRows) {
			hitRows = append(hitRows, val.Row)
		}
	}
	return len(hitRows)
}

//Returns # of cols with at least 1 true value
func (sm *SparseBinaryMatrix) TotalTrueCols() int {
	var hitCols []int
	for _, val := range sm.entries {
		if !utils.ContainsInt(val.Col, hitCols) {
			hitCols = append(hitCols, val.Col)
		}
	}
	return len(hitCols)
}

//Returns total true entries
func (sm *SparseBinaryMatrix) TotalNonZeroCount() int {
	return len(sm.entries)
}

// Ors 2 matrices
func (sm *SparseBinaryMatrix) Or(sm2 *SparseBinaryMatrix) *SparseBinaryMatrix {
	result := NewSparseBinaryMatrix(sm.Height, sm.Width)

	for _, val := range sm.entries {
		result.Set(val.Row, val.Col, true)
	}

	for _, val := range sm2.entries {
		result.Set(val.Row, val.Col, true)
	}

	return result
}

//Clears  all entries
func (sm *SparseBinaryMatrix) Clear() {
	sm.entries = nil
}

//Fills specified row with specified value
func (sm *SparseBinaryMatrix) FillRow(row int, val bool) {
	for j := 0; j < sm.Width; j++ {
		sm.Set(row, j, val)
	}
}

//Copys a matrix
func (sm *SparseBinaryMatrix) Copy() *SparseBinaryMatrix {
	if sm == nil {
		return nil
	}

	result := new(SparseBinaryMatrix)
	result.Width = sm.Width
	result.Height = sm.Height
	result.entries = make([]SparseEntry, len(sm.entries))
	for idx, val := range sm.entries {
		result.entries[idx] = val
	}

	return result
}

func (sm *SparseBinaryMatrix) ToString() string {
	var buffer bytes.Buffer

	for r := 0; r < sm.Height; r++ {
		for c := 0; c < sm.Width; c++ {
			if sm.Get(r, c) {
				buffer.WriteByte('1')
			} else {
				buffer.WriteByte('0')
			}
		}
		buffer.WriteByte('\n')
	}

	return buffer.String()
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
