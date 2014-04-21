package htm

//Items are index of non-zero columns
type SparseRow []int

//Sparse binary matrix stores indexes of non-zero entries in matrix
//to conserve space
//Rows is map of non-zero rows indexed by row index
type SparseBinaryMatrix struct {
	Cols              int
	Rows              int
	TotalNonZeroCount int
	//NonZeroRows       []int
	Rows map[int][]SparseRow
}

func NewSparseBinaryMatrix(width int, height int) SparseBinaryMatrix {
	m := SparseBinaryMatrix{}
	m.Rows = height
	m.Cols = width
	m.Rows = make(map[int][]SparseRow, height*.03)

	return m
}

func NewRandSparseBinaryMatrix() *SparseBinaryMatrix {

}

func (sm *SparseBinaryMatrix) Resize(width int, height int) {

}

//Get value at col,row position
func (sm *SparseBinaryMatrix) Get(col int, row int) bool {
	if r, ok := sm.Rows[row]; ok {
		for x := 0; x < len(r); x++ {
			if r[x] == col {
				return true
			}
		}
	}
	return false
}

func (sm *SparseBinaryMatrix) delete(col int, row int) {
	if r, ok := sm.Rows[row]; ok {
		for x := 0; x < len(r); x++ {
			if r[x] == col {
				sm.Rows[row] = append(r[:x], r[x+1:])
				break
			}
		}
		if len(sm.Rows[row]) < 1 {
			//delete row entry
			delete(sm.Rows, row)
		}
	}
}

//Set value at col,row position
func (sm *SparseBinaryMatrix) Set(col int, row int, value bool) {
	if value {
		if r, ok := sm.Rows[row]; ok {
			sm.Rows[row] = append(sm.Rows[row], col)
		} else {
			sm.Rows[row] = SparseRow{col}
		}
	} else {
		sm.delete(col, row)
	}

}
