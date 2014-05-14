package htm

import (
	//"fmt"
	"math/big"
)

type TupleInt struct {
	A int
	B int
}

//Euclidean modulous
func Mod(a, b int) int {
	ab := big.NewInt(int64(a))
	bb := big.NewInt(int64(b))
	return int(ab.Mod(ab, bb).Int64())
}

//Dot product
func DotInt(a, b []int) int {
	if len(a) != len(b) {
		panic("Params have differing lengths")
	}
	result := 0
	for i := range a {
		result += a[i] * b[i]
	}
	return result
}

//Populates integer slice with index values
func FillSliceInt(values []int) {
	for i := range values {
		values[i] = i
	}
}

//Returns cartesian product of specified
//2d arrayb
func CartProductInt(values [][]int) [][]int {
	pos := make([]int, len(values))
	var result [][]int

	for pos[0] < len(values[0]) {
		temp := make([]int, len(values))
		for j := 0; j < len(values); j++ {
			temp[j] = values[j][pos[j]]
		}
		result = append(result, temp)
		pos[len(values)-1]++
		for k := len(values) - 1; k >= 1; k-- {
			if pos[k] >= len(values[k]) {
				pos[k] = 0
				pos[k-1]++
			} else {
				break
			}
		}
	}
	return result
}

//Searches int slice for specified integer
func ContainsInt(q int, vals []int) bool {
	for _, val := range vals {
		if val == q {
			return true
		}
	}
	return false
}
