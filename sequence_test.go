package htm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitOnes(t *testing.T) {
	seq := Ones(5)

	assert.Equal(t, 5, seq.Len())
	assert.Equal(t, []bool{true, true, true, true, true}, seq.ToSlice())

}

func TestInit(t *testing.T) {
	seq := BinarySequence(5)

	assert.Equal(t, 5, seq.Len())
	assert.Equal(t, []bool{false, false, false, false, false}, seq.ToSlice())

}

func TestInitFromInt(t *testing.T) {
	seq := FromInts([]int{0, 0, 1, 0, 0, 1, 0})

	assert.Equal(t, 7, seq.Len())
	assert.Equal(t, "0010010", seq.String())
	t.Log(seq.String())
	assert.Equal(t, []bool{false, false, true, false, false, true, false}, seq.ToSlice())

}

func TestInitFromStr(t *testing.T) {
	seq := FromStr("0010010")

	assert.Equal(t, 7, seq.Len())
	assert.Equal(t, []bool{false, false, true, false, false, true, false}, seq.ToSlice())

}

func TestLen(t *testing.T) {
	seq := FromStr("0010010")

	assert.Equal(t, 7, seq.Len())

	seq = FromStr("0010010000001001000000100100000010010000001001000000100100000010010000000000000")

	assert.Equal(t, 79, seq.Len())
}

func TestOnBits(t *testing.T) {
	seq := FromStr("0010010")

	assert.Equal(t, 2, len(seq.OnIndices()))

	seq = FromStr("0010010000001001000000100100000010010000001001000000100100000010010000000000000")

	assert.Equal(t, 14, len(seq.OnIndices()))
}

func TestSet(t *testing.T) {
	seq := BinarySequence(5)

	seq.Set(0, true)
	seq.Set(4, true)
	seq.Set(2, true)

	assert.Equal(t, []bool{true, false, true, false, true}, seq.ToSlice())
}

func TestGet(t *testing.T) {
	seq := BinarySequence(5)

	seq.Set(0, true)
	seq.Set(4, true)
	seq.Set(2, true)

	assert.True(t, seq.At(0))
	assert.True(t, seq.At(4))
	assert.True(t, seq.At(2))
	assert.False(t, seq.At(1))
	assert.False(t, seq.At(3))
}

func TestToSlice(t *testing.T) {
	seq := BinarySequence(10)
	s := seq.ToSlice()

	assert.Equal(t, make([]bool, 10), s)

	seq = Ones(5)
	s = seq.ToSlice()

	assert.Equal(t, []bool{true, true, true, true, true}, s)
}

func TestEquals(t *testing.T) {

	a := Ones(6)
	b := Ones(6)

	assert.True(t, a.Equals(b))

	c := BinarySequence(6)

	assert.False(t, a.Equals(c))
	assert.True(t, a.Equals(b))

	b = Ones(3)

	assert.False(t, a.Equals(b))
	assert.False(t, c.Equals(b))
}

func TestOnIndices(t *testing.T) {
	seq := Ones(5)

	assert.Equal(t, 5, len(seq.OnIndices()))

	seq = BinarySequence(5)

	assert.Equal(t, 0, len(seq.OnIndices()))

}

func TestOr(t *testing.T) {

	a := Ones(5)
	b := BinarySequence(5)

	result := a.Or(b)

	assert.True(t, Ones(5).Equals(result))

	c := BinarySequence(5)
	c.Set(2, true)
	c.Set(4, true)

	result = b.Or(c)

	assert.Equal(t, []bool{false, false, true, false, true}, result.ToSlice())

}

func TestAnd(t *testing.T) {

	a := Ones(5)
	b := BinarySequence(5)

	result := a.And(b)

	assert.True(t, BinarySequence(5).Equals(result))

	c := BinarySequence(5)
	c.Set(2, true)
	c.Set(4, true)

	result = a.And(c)

	assert.Equal(t, []bool{false, false, true, false, true}, result.ToSlice())

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

	seq.Set(4, true)
	assert.False(t, seq.Contains(subSeq))
	seq.Set(5, true)
	assert.False(t, seq.Contains(subSeq))
	seq.Set(9, true)
	assert.False(t, seq.Contains(subSeq))
	seq.Set(6, true)
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

	a.Append(b)

	assert.Equal(t, x+y, a.String())

}

func TestCopy(t *testing.T) {

	str := "1010101010010001"

	seq := FromStr(str)

	result := seq.Copy()

	assert.Equal(t, str, result.String())

}
