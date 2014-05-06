package htm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCartProductInt(t *testing.T) {
	vals := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{10, 11, 12, 13},
	}

	result := CartProductInt(vals)

	assert.Equal(t, 48, len(result))
	assert.Equal(t, []int{1, 5, 10}, result[0])
	assert.Equal(t, []int{2, 5, 12}, result[18])
	assert.Equal(t, []int{3, 8, 13}, result[47])

}
