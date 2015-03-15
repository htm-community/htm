package htm

import (
	"bytes"
	"fmt"
	"github.com/cznic/mathutil"
	"math"
)

type Sequence struct {
	data         []uint64
	binaryLength int
	onBits       int
}

/* Intializers */

func BinarySequence(length int) *Sequence {
	seq := new(Sequence)
	seq.init(length)
	return seq
}

func FromInts(ints []int) *Sequence {
	seq := new(Sequence)
	seq.init(len(ints))
	for idx, val := range ints {
		if val > 0 {
			seq.Set(idx, true)
		}
	}
	return seq
}

func FromStr(str string) *Sequence {
	seq := new(Sequence)
	seq.init(len(str))
	for idx, val := range str {
		if val != '0' {
			seq.Set(idx, true)
		}
	}
	return seq
}

func Ones(length int) *Sequence {
	seq := new(Sequence)
	seq.init(length)
	for idx, _ := range seq.data {
		seq.data[idx] = math.MaxUint64
	}
	return seq
}

/* helpers */

func (s *Sequence) init(size int) {
	newSize, _ := s.idx(size)
	newSize += 1
	s.data = make([]uint64, newSize)
	s.binaryLength = size
}

//returns slice and bit position of binary index
func (s *Sequence) idx(i int) (int, int) {
	return i / 64, i % 64
}

/* exported functions */

func (s *Sequence) Equals(other *Sequence) bool {
	if s.Len() != other.Len() {
		return false
	}

	for idx, val := range s.data {
		if val != other.data[idx] {
			return false
		}
	}

	return true
}

/*
	appends a sequence (causes new allocation)
*/
func (s *Sequence) Append(other *Sequence) {
	newData := make([]uint64, len(s.data)+len(other.data))
	copy(newData, s.data)
	s.data = newData

	for i := 0; i < other.Len(); i++ {
		s.Set(i+s.binaryLength, other.At(i))
	}

	s.binaryLength += other.Len()
}

/*
 Or's 2 binary sequences starting from 0 index
*/
func (s *Sequence) Or(other *Sequence) *Sequence {
	length := mathutil.Max(s.Len(), other.Len())
	result := BinarySequence(length)

	bound := mathutil.Min(len(s.data), len(other.data))
	for i := 0; i < bound; i++ {
		result.data[i] = s.data[i] | other.data[i]
	}

	return result
}

/*
	And's 2 binary sequences starting from 0 index
*/
func (s *Sequence) And(other *Sequence) *Sequence {
	length := mathutil.Max(s.Len(), other.Len())
	result := BinarySequence(length)

	bound := mathutil.Min(len(s.data), len(other.data))
	for i := 0; i < bound; i++ {
		result.data[i] = s.data[i] & other.data[i]
	}

	return result
}

func (s *Sequence) At(idx int) bool {
	pos, bitPos := s.idx(idx)
	return !(s.data[pos]&1<<uint(bitPos) == 0)
}

func (s *Sequence) Set(idx int, val bool) {
	pos, bitPos := s.idx(idx)
	if val {
		s.data[pos] |= (1 << uint(bitPos))
	} else {
		s.data[pos] &= ^(1 << uint(bitPos))
	}

}

/*
	Length of binary sequence
*/
func (s *Sequence) Len() int {
	return s.binaryLength
}

/*
	Set value of specified indices
*/
func (s *Sequence) SetIndices(indices []int, val bool) {
	for _, v := range indices {
		s.Set(v, val)
	}
}

/*
	Returns the indices of all on bits
*/
func (s *Sequence) OnIndices() []int {
	result := make([]int, 0, s.Len()/3)

	for i := 0; i < s.Len(); i++ {
		if s.At(i) {
			result = append(result, i)
		}
	}

	return result
}

/*
	Returns true if binary sequence contains specified subsequence
*/
func (s *Sequence) Contains(subSequence *Sequence) bool {
	if subSequence.Len() > s.Len() {
		return false
	}

	idx := 0
	subIdx := 0
	for idx < s.Len() {
		if s.data[idx] != subSequence.data[subIdx] {
			subIdx = 0
		} else {
			subIdx++
		}

		if subIdx >= len(subSequence.data) {
			return true
		}

		idx++
	}

	return false
}

/*
	returns string representation of the binary sequence
*/
func (s *Sequence) String() string {
	var buffer bytes.Buffer
	for _, val := range s.data {
		buffer.WriteString(fmt.Sprintf("%b", val))
	}

	return buffer.String()[:s.binaryLength]
}

/*
	returns slice of bits
*/
func (s *Sequence) ToSlice() []bool {
	result := make([]bool, s.Len())
	for bit := 0; bit < s.Len(); bit++ {
		result = append(result, s.At(bit))
	}
	return result
}

/*
	returns copy of sequence
*/
func (s *Sequence) Copy() *Sequence {
	cpy := BinarySequence(s.binaryLength)
	copy(cpy.data, s.data)
	return cpy
}
