package encoders

import (
	"github.com/stretchr/testify/assert"
	"github.com/zacg/htm/utils"
	"testing"
)

func TestOrderForCoord(t *testing.T) {
	h1 := order([]int{2, 5, 10})
	h2 := order([]int{2, 5, 11})
	h3 := order([]int{2497477, -923478})

	assert.True(t, 0 <= h1 && h1 < 1)

	t.Log(h1)
	t.Log(h2)
	assert.True(t, h1 != h2)
	assert.True(t, h2 != h3)

}

func TestBasicEncode(t *testing.T) {
	e := NewCoordinateEncoder(5, 33)
	output := make([]bool, 33)
	input := []string{"100,200", "7"}

	e.Encode(input, output)
	t.Log("output:", output)

	assert.Equal(t, 5, utils.CountTrue(output))

	output2 := make([]bool, 33)
	e.Encode(input, output2)
	assert.Equal(t, output, output2)

}
