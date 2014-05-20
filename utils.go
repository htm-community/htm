package htm

import (
	//"fmt"
	"math"
	"math/big"
	"math/rand"
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

//Creates an integer slice with indices containing
// the specified initial value
func MakeSliceInt(size, initialValue int) []int {
	result := make([]int, size)
	if initialValue != 0 {
		for i, _ := range result {
			result[i] = initialValue
		}
	}
	return result
}

func MakeSliceFloat64(size int, initialValue float64) []float64 {
	result := make([]float64, size)
	if initialValue != 0 {
		for i, _ := range result {
			result[i] = initialValue
		}
	}
	return result
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

func ContainsFloat64(q float64, vals []float64) bool {
	for _, val := range vals {
		if val == q {
			return true
		}
	}
	return false
}

type CompareInt func(int) bool

func CountInt(q CompareInt, vals []int) int {
	count := 0
	for i := range vals {
		if q(i) {
			count++
		}
	}
	return count
}

func randFloatRange(min, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

//returns max index wise comparison
func MaxInt(a, b []int) []int {
	result := make([]int, len(a))
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			result[i] = a[i]
		} else {
			result[i] = b[i]
		}
	}

	return result
}

//Returns max value from specified int slice
func MaxSliceInt(values []int) int {
	max := 0
	for i := 0; i < len(values); i++ {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}

//Returns max value from specified float slice
func MaxSliceFloat64(values []float64) float64 {
	max := 0.0
	for i := 0; i < len(values); i++ {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}

//Returns product of set of integers
func ProdInt(vals []int) int {
	sum := 1
	for x := 0; x < len(vals); x++ {
		sum *= vals[x]
	}

	if sum == 1 {
		return 0
	} else {
		return sum
	}
}

//Returns cumulative product
func CumProdInt(vals []int) []int {
	if len(vals) < 2 {
		return vals
	}
	result := make([]int, len(vals))
	result[0] = vals[0]
	for x := 1; x < len(vals); x++ {
		result[x] = vals[x] * result[x-1]
	}

	return result
}

//Returns cumulative product starting from end
func RevCumProdInt(vals []int) []int {
	if len(vals) < 2 {
		return vals
	}
	result := make([]int, len(vals))
	result[len(vals)-1] = vals[len(vals)-1]
	for x := len(vals) - 2; x >= 0; x-- {
		result[x] = vals[x] * result[x+1]
	}

	return result
}

func RoundPrec(x float64, prec int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}

	sign := 1.0
	if x < 0 {
		sign = -1
		x *= -1
	}

	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)

	if frac >= 0.5 {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow * sign
}
