package htm

import (
	"github.com/nupic-community/htm/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInitOnes(t *testing.T) {
	seq := Ones(5)

	assert.Equal(t, 5, seq.Len())
	assert.Equal(t, []bool{true, true, true, true, true}, seq.Slice())

}

func TestInit(t *testing.T) {
	seq := BinarySequence(5)

	assert.Equal(t, 5, seq.Len())
	assert.Equal(t, []bool{false, false, false, false, false}, seq.Slice())

}

func TestInitFromInt(t *testing.T) {
	seq := FromInts([]int{0, 0, 1, 0, 0, 1, 0})

	assert.Equal(t, 7, seq.Len())
	assert.Equal(t, []bool{false, false, true, false, false, true, false}, seq.Slice())

}

func TestInitFromStr(t *testing.T) {
	seq := FromStr("0010010")

	assert.Equal(t, 7, seq.Len())
	assert.Equal(t, []bool{false, false, true, false, false, true, false}, seq.Slice())

}

func TestLen(t *testing.T) {
	seq := FromStr("0010010")

	assert.Equal(t, 7, seq.Len())

	seq = FromStr("0010010000001001000000100100000010010000001001000000100100000010010000000000000")

	assert.Equal(t, 79, seq.Len())
}

func TestOnBits(t *testing.T) {
	seq := FromStr("0010010")

	assert.Equal(t, 2, seq.OnBits())

	seq = FromStr("0010010000001001000000100100000010010000001001000000100100000010010000000000000")

	assert.Equal(t, 14, seq.OnBits())
}

func TestSet(t *testing.T) {
	seq := BinarySequence(5)

	seq.Set(0, true)
	seq.Set(4, true)
	seq.Set(2, true)

	assert.Equal(t, []bool{true, false, true, false, true}, seq.Slice())
}

func TestGet(t *testing.T) {
	seq := BinarySequence(5)

	seq.Set(0, true)
	seq.Set(4, true)
	seq.Set(2, true)

	assert.True(t, seq.Get(0))
	assert.True(t, seq.Get(4))
	assert.True(t, seq.Get(2))
	assert.False(t, seq.Get(1))
	assert.False(t, seq.Get(3))
}

func TestToSlice(t *testing.T) {
	seq := BinarySequence(10)
	s := seq.Slice()

	assert.Equal(t, make([]bool, 10), s)

	seq = Ones(5)
	s = seq.Slice()

	assert.Equal(t, []bool{true, true, true, true, true}, s)
}

func TestEquals(t *testing.T) {

	a := Ones(6)
	b := Ones(6)

	assert.True(t, a.Equals(b))

	c := BinarySequence(6)

	assert.False(t, a.Equals(c))
	assert.False(t, a.Equals(b))

	b = Ones(3)

	assert.False(t, a.Equals(b))

}

func TestOnIndices(t *testing.T) {
	seq := Ones(5)

	assert.Equal(t, 5, seq.OnIndices())

	seq = BinarySequence(5)

	assert.Equal(t, 0, seq.OnIndices())

}

func TestOr(t *testing.T) {

	a := Ones(5)
	b := BinarySequence(5)

	result := a.Or(b)

	assert.True(t, Ones(5).Equals(result))

	c := BinarySequence(5)
	c.Set(2, 0)
	c.Set(4, 0)

	result = b.Or(c)

	assert.Equal(t, []bool{false, false, true, false, true}, result.Slice())

}

func TestAnd(t *testing.T) {

	a := Ones(5)
	b := BinarySequence(5)

	result := a.And(b)

	assert.True(t, BinarySequence(5).Equals(result))

	c := BinarySequence(5)
	c.Set(2, false)
	c.Set(4, false)

	result = a.And(c)

	assert.Equal(t, []bool{false, false, true, false, true}, result.Slice())

}

func TestToString(t *testing.T) {
	str := "11000110101010"
	seq := FromStr(str)

	assert.Equal(t, str, seq.String())

}

func TestContains(t *testing.T) {

	seq := BinarySequence(10)
	subSeq := Ones(3)

	assert.False(t, seq.Contains(subSeq))

	seq.set(4, true)
	assert.False(t, seq.Contains(subSeq))
	seq.set(5, true)
	assert.False(t, seq.Contains(subSeq))
	seq.set(9, true)
	assert.False(t, seq.Contains(subSeq))
	seq.set(6, true)
	assert.True(t, seq.Contains(subSeq))

}

func TestSetIndices(t *testing.T) {

	seq := BinarySequence(10)

	seq.SetIndices([]int{8, 7, 2}, true)

	assert.Equal(t, "0010000110", seq.String())

}

func TestAppend(t *testing.T) {

	x := "111001010101"
	y := "1010101010010001"

	a := FromStr(x)
	b := FromStr(y)

	result := a.Append(b)

	assert.Equal(t, x+y, result.String())

}
