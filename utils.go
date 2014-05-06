package htm

import (
//"fmt"
)

type TupleInt struct {
	A int
	B int
}

//Returns cartesian product of specified
//2d arrayb
func CartProductInt(values [][]int) [][]int {
	pos := make([]int, len(values))
	var result [][]int

	for pos[0] < len(values[0])-1 {
		temp := make([]int, len(values))
		for j := 0; j < len(values); j++ {
			temp[j] = values[j][pos[j]]
		}
		result = append(result, temp)
		pos[len(values)-1]++
		for k := len(values) - 1; k >= 0; k-- {
			if pos[k] >= len(values[k]) {
				pos[k] = 0
				pos[k-1]++
			}
		}
	}
	return result
}
