package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFillSliceWithIdxInt(t *testing.T) {
	vals := make([]int, 3)
	FillSliceWithIdxInt(vals)
	expected := []int{0, 1, 2}
	assert.Equal(t, expected, vals)
}

func TestCartProductInt(t *testing.T) {
	vals := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{10, 11, 12, 13},
	}

	result := CartProductInt(vals)

	assert.Equal(t, 64, len(result))
	assert.Equal(t, []int{1, 5, 10}, result[0])
	assert.Equal(t, []int{2, 5, 12}, result[18])
	assert.Equal(t, []int{3, 8, 13}, result[47])

	vals = [][]int{
		{1, 2},
		{2, 3},
		{0, 1},
	}

	result = CartProductInt(vals)

	assert.Equal(t, 8, len(result))

}

func TestProdInt(t *testing.T) {

	vals := []int{32, 32}
	expected := 1024

	actual := ProdInt(vals)

	assert.Equal(t, expected, actual)

}
