package encoders

import (
	"github.com/stretchr/testify/assert"
	"github.com/zacg/htm/utils"
	"testing"
)

func TestSimpleEncoding(t *testing.T) {

	p := NewScalerEncoderParams(3, 1, 8)
	//p.Resolution = 1
	p.N = 14
	//p.Width = 3
	//p.MaxVal = 8
	//p.MinVal = 1
	p.Periodic = true
	//p.Verbosity = 5

	e := NewScalerEncoder(p)

	encoded := e.Encode(1, false)
	t.Log(encoded)
	expected := utils.Make1DBool([]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	assert.True(t, len(encoded) == 14)
	assert.Equal(t, expected, encoded)

	encoded = e.Encode(2, false)
	expected = utils.Make1DBool([]int{0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.True(t, len(encoded) == 14)
	assert.Equal(t, expected, encoded)

	encoded = e.Encode(3, false)
	expected = utils.Make1DBool([]int{0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.True(t, len(encoded) == 14)
	assert.Equal(t, expected, encoded)

}
