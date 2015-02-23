package htm

import (
	"bytes"
	"github.com/nupic-community/htmutils"
	//"math"
)

//Sparse binary matrix stores indexes of non-zero entries in matrix
//to conserve space
type DenseBinaryMatrix struct {
	Width   int
	Height  int
	entries []bool
}

//Create new sparse binary matrix of specified size
func NewDenseBinaryMatrix(height, width int) *DenseBinaryMatrix {
	m := &DenseBinaryMatrix{}
	m.Height = height
	m.Width = width
	m.entries = make([]bool, width*height)
	return m
}

//Create sparse binary matrix from specified dense matrix
func NewDenseBinaryMatrixFromDense(values [][]bool) *DenseBinaryMatrix {
	if len(values) < 1 {
		panic("No values specified.")
	}

	m := NewDenseBinaryMatrix(len(values), len(values[0]))
	for r := 0; r < m.Height; r++ {
		m.SetRowFromDense(r, values[r])
	}
	return m
}

//Create sparse binary matrix from specified dense matrix
func NewDenseBinaryMatrixFromDense1D(values []bool, rows, cols int) *DenseBinaryMatrix {
	if len(values) < 1 {
		panic("No values specified.")
	}
	if len(values) != rows*cols {
		panic("Invalid size")
	}

	m := NewDenseBinaryMatrix(rows, cols)

	for r := 0; r < m.Height; r++ {
		m.SetRowFromDense(r, values[r*cols:(r*cols)+cols])
	}

	return m
}

// Creates a sparse binary matrix from specified integer array
// (any values greater than 0 are true)
func NewDenseBinaryMatrixFromInts(values [][]int) *DenseBinaryMatrix {
	if len(values) < 1 {
		panic("No values specified.")
	}

	m := NewDenseBinaryMatrix(len(values), len(values[0]))

	for r := 0; r < m.Height; r++ {
		for c := 0; c < m.Width; c++ {
			if values[r][c] > 0 {
				m.Set(r, c, true)
			}
		}
	}

	return m
}

//Converts index to col/row
func (sm *DenseBinaryMatrix) toIndex(index int) (row int, col int) {
	row = index / sm.Width
	col = index % sm.Width
	return
}

//Returns all true/on indices
func (sm *DenseBinaryMatrix) Entries() []SparseEntry {
	result := make([]SparseEntry, 0, int(float64(len(sm.entries))*0.3))
	for idx, val := range sm.entries {
		if val {
			i, j := sm.toIndex(idx)
			result = append(result, SparseEntry{i, j})
		}
	}
	return result
}

//Returns flattend dense represenation
func (sm *DenseBinaryMatrix) Flatten() []bool {
	result := make([]bool, sm.Height*sm.Width)
	for _, val := range sm.Entries() {
		result[(val.Row*sm.Width)+val.Col] = true
	}
	return result
}

//Get value at col,row position
func (sm *DenseBinaryMatrix) Get(row int, col int) bool {
	row = row % sm.Height
	if row < 0 {
		row = sm.Height - row
	}
	col = col % sm.Width
	if col < 0 {
		col = sm.Width - col
	}

	return sm.entries[row*sm.Width+col]
}

//Set value at row,col position
func (sm *DenseBinaryMatrix) Set(row int, col int, value bool) {
	row = row % sm.Height
	if row < 0 {
		row = sm.Height - row
	}
	col = col % sm.Width
	if col < 0 {
		col = sm.Width - col
	}
	sm.entries[row*sm.Width+col] = value
}

//Replaces specified row with values, assumes values is ordered
//correctly
func (sm *DenseBinaryMatrix) ReplaceRow(row int, values []bool) {
	sm.validateRowCol(row, len(values))

	for i := 0; i < sm.Width; i++ {
		sm.Set(row, i, values[i])
	}
}

//Replaces row with true values at specified indices
func (sm *DenseBinaryMatrix) ReplaceRowByIndices(row int, indices []int) {
	sm.validateRow(row)
	start := row * sm.Width
	for i := 0; i < sm.Width; i++ {
		sm.entries[start+i] = utils.ContainsInt(i, indices)
	}
}

//Returns dense row
func (sm *DenseBinaryMatrix) GetDenseRow(row int) []bool {
	sm.validateRow(row)
	result := make([]bool, sm.Width)

	start := row * sm.Width
	for i := 0; i < sm.Width; i++ {
		result[i] = sm.entries[start+i]
	}

	return result
}

//Returns a rows "on" indices
func (sm *DenseBinaryMatrix) GetRowIndices(row int) []int {
	result := make([]int, 0, sm.Width)
	start := row * sm.Width
	for i := 0; i < sm.Width; i++ {
		if sm.entries[start+i] {
			result = append(result, i)
		}
	}
	return result
}

//Sets a sparse row from dense representation
func (sm *DenseBinaryMatrix) SetRowFromDense(row int, denseRow []bool) {
	//TODO: speed this up
	sm.validateRowCol(row, len(denseRow))
	for i := 0; i < sm.Width; i++ {
		sm.Set(row, i, denseRow[i])
	}
}

//In a normal matrix this would be multiplication in binary terms
//we just and then sum the true entries
func (sm *DenseBinaryMatrix) RowAndSum(row []bool) []int {
	sm.validateCol(len(row))
	result := make([]int, sm.Height)

	for idx, val := range sm.entries {
		if val {
			r, c := sm.toIndex(idx)
			if row[c] {
				result[r]++
			}
		}
	}

	return result
}

//Returns row indexes with at least 1 true column
func (sm *DenseBinaryMatrix) NonZeroRows() []int {
	counts := make(map[int]int, sm.Height)

	for idx, val := range sm.entries {
		if val {
			r, _ := sm.toIndex(idx)
			counts[r]++

		}
	}

	result := make([]int, 0, sm.Height)
	for k, v := range counts {
		if v > 0 && !utils.ContainsInt(k, result) {
			result = append(result, k)
		}
	}
	return result
}

//Returns # of rows with at least 1 true value
func (sm *DenseBinaryMatrix) TotalTrueRows() int {
	return len(sm.NonZeroRows())
}

//Returns total true entries
func (sm *DenseBinaryMatrix) TotalNonZeroCount() int {
	return len(sm.Entries())
}

// Ors 2 matrices
func (sm *DenseBinaryMatrix) Or(sm2 *DenseBinaryMatrix) *DenseBinaryMatrix {
	result := NewDenseBinaryMatrix(sm.Height, sm.Width)

	for _, val := range sm.Entries() {
		result.Set(val.Row, val.Col, true)
	}

	for _, val := range sm2.Entries() {
		result.Set(val.Row, val.Col, true)
	}

	return result
}

//Clears  all entries
func (sm *DenseBinaryMatrix) Clear() {
	utils.FillSliceBool(sm.entries, false)
}

//Fills specified row with specified value
func (sm *DenseBinaryMatrix) FillRow(row int, val bool) {
	for j := 0; j < sm.Width; j++ {
		sm.Set(row, j, val)
	}
}

//Copys a matrix
func (sm *DenseBinaryMatrix) Copy() *DenseBinaryMatrix {
	if sm == nil {
		return nil
	}

	result := new(DenseBinaryMatrix)
	result.Width = sm.Width
	result.Height = sm.Height
	result.entries = make([]bool, len(sm.entries))
	for idx, val := range sm.entries {
		result.entries[idx] = val
	}

	return result
}

func (sm *DenseBinaryMatrix) ToString() string {
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

func (sm *DenseBinaryMatrix) validateCol(col int) {
	if col > sm.Width {
		panic("Specified row is wider than matrix.")
	}
}

func (sm *DenseBinaryMatrix) validateRow(row int) {
	if row > sm.Height {
		panic("Specified row is out of bounds.")
	}
}

func (sm *DenseBinaryMatrix) validateRowCol(row int, col int) {
	sm.validateCol(col)
	sm.validateRow(row)
}
