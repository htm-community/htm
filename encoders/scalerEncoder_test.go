package encoders

import (
	"github.com/nupic-community/htm/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleEncoding(t *testing.T) {

	p := NewScalerEncoderParams(3, 1, 8)
	p.N = 14
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

func TestWideEncoding(t *testing.T) {

	p := NewScalerEncoderParams(5, 0, 24)
	p.Periodic = true
	//p.Verbosity = 5
	p.Radius = 4
	e := NewScalerEncoder(p)

	encoded := e.Encode(14.916666666666666, false)
	t.Log(encoded)
	expected := utils.Make1DBool([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	assert.True(t, len(encoded) == 30)
	assert.Equal(t, utils.Bool2Int(expected), utils.Bool2Int(encoded))

}

func TestNarrowEncoding(t *testing.T) {

	p := NewScalerEncoderParams(3, 0, 1)
	p.Periodic = false
	//p.Verbosity = 5
	p.Radius = 1
	e := NewScalerEncoder(p)

	encoded := make([]bool, 6)
	e.EncodeToSlice(0, false, encoded)
	t.Log(encoded)
	expected := utils.Make1DBool([]int{1, 1, 1, 0, 0, 0})

	assert.True(t, len(encoded) == 6)
	assert.Equal(t, utils.Bool2Int(expected), utils.Bool2Int(encoded))

}

func TestSimpleDecoding(t *testing.T) {

	p := NewScalerEncoderParams(3, 1, 8)
	p.Radius = 1.5
	p.Periodic = true
	//p.Verbosity = 5

	e := NewScalerEncoder(p)

	// Test with a "hole"
	encoded := utils.Make1DBool([]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0})
	expected := []utils.TupleFloat{utils.TupleFloat{7.5, 7.5}}
	actual := e.Decode(encoded)
	assert.Equal(t, expected, actual)

	// Test with something wider than w, and with a hole, and wrapped
	encoded = utils.Make1DBool([]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0})
	expected = []utils.TupleFloat{utils.TupleFloat{7.5, 8}, utils.TupleFloat{1, 1}}
	actual = e.Decode(encoded)
	assert.Equal(t, expected, actual)

	// Test with something wider than w, no hole
	encoded = utils.Make1DBool([]int{1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	expected = []utils.TupleFloat{utils.TupleFloat{1.5, 2.5}}
	actual = e.Decode(encoded)
	assert.Equal(t, expected, actual)

	// 1
	encoded = utils.Make1DBool([]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
	expected = []utils.TupleFloat{utils.TupleFloat{1, 1}}
	actual = e.Decode(encoded)
	assert.Equal(t, expected, actual)

	// 2
	encoded = utils.Make1DBool([]int{0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	expected = []utils.TupleFloat{utils.TupleFloat{2, 2}}
	actual = e.Decode(encoded)
	assert.Equal(t, expected, actual)

}
