package utils

import (
	//"fmt"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"time"
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
func FillSliceWithIdxInt(values []int) {
	for i := range values {
		values[i] = i
	}
}

//Populates float64 slice with specified value
func FillSliceInt(values []int, value int) {
	for i := range values {
		values[i] = value
	}
}

//Populates float64 slice with specified value
func FillSliceFloat64(values []float64, value float64) {
	for i := range values {
		values[i] = value
	}
}

//Populates bool slice with specified value
func FillSliceBool(values []bool, value bool) {
	for i := range values {
		values[i] = value
	}
}

//Populates bool slice with specified value
func FillSliceRangeBool(values []bool, value bool, start, length int) {
	for i := 0; i < length; i++ {
		values[start+i] = value
	}
}

//Returns the subset of values specified by indices
func SubsetSliceInt(values, indices []int) []int {
	result := make([]int, len(indices))
	for i, val := range indices {
		result[i] = values[val]
	}
	return result
}

//Returns the subset of values specified by indices
func SubsetSliceFloat64(values []float64, indices []int) []float64 {
	result := make([]float64, len(indices))
	for i, val := range indices {
		result[i] = values[val]
	}
	return result
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

// type CompareInt func(int) bool

// func CountInt(q CompareInt, vals []int) int {
// 	count := 0
// 	for i := range vals {
// 		if q(i) {
// 			count++
// 		}
// 	}
// 	return count
// }

func RandFloatRange(min, max float64) float64 {
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

//Helper for unit tests where int literals are easier
// to read
func Make2DBool(values [][]int) [][]bool {
	result := make([][]bool, len(values))

	for i, val := range values {
		result[i] = make([]bool, len(val))
		for j, col := range val {
			result[i][j] = col == 1
		}
	}

	return result
}

func Make1DBool(values []int) []bool {
	result := make([]bool, len(values))
	for i, val := range values {
		result[i] = val == 1
	}
	return result
}

//Returns number of on bits
func CountInt(values []int, value int) int {
	count := 0
	for _, val := range values {
		if val == value {
			count++
		}
	}
	return count
}

//Returns number of on bits
func CountFloat64(values []float64, value float64) int {
	count := 0
	for _, val := range values {
		if val == value {
			count++
		}
	}
	return count
}

//Returns number of on bits
func CountTrue(values []bool) int {
	count := 0
	for _, val := range values {
		if val {
			count++
		}
	}
	return count
}

//Or's 2 bool slices
func OrBool(a, b []bool) []bool {
	result := make([]bool, len(a))
	for i, val := range a {
		result[i] = val || b[i]
	}
	return result
}

//Returns random slice of floats of specified length
func RandomSample(length int) []float64 {
	result := make([]float64, length)

	for i, _ := range result {
		result[i] = rand.Float64()
	}

	return result
}

func Bool2Int(s []bool) []int {
	result := make([]int, len(s))
	for idx, val := range s {
		if val {
			result[idx] = 1
		} else {
			result[idx] = 0
		}

	}
	return result
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n", name, elapsed)
}

func SumSliceFloat64(values []float64) float64 {
	result := 0.0
	for _, val := range values {
		result += val
	}
	return result
}

//Returns "on" indices
func OnIndices(s []bool) []int {
	var result []int
	for idx, val := range s {
		if val {
			result = append(result, idx)
		}
	}
	return result
}

// Returns complement of s and t
func Complement(s []int, t []int) []int {
	result := make([]int, 0, len(s))
	for _, val := range s {
		found := false
		for _, v2 := range t {
			if v2 == val {
				found = true
				break
			}
		}
		if !found {
			result = append(result, val)
		}
	}
	return result
}

func Add(s []int, t []int) []int {
	result := make([]int, 0, len(s)+len(t))
	result = append(result, s...)

	for _, val := range t {
		if !ContainsInt(val, s) {
			result = append(result, val)
		}
	}
	return result
}
