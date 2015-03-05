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
