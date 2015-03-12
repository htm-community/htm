package htm

import (
	"fmt"
	//"github.com/cznic/mathutil"
	//"github.com/zacg/floats"
	//"github.com/nupic-community/htm"
	"github.com/nupic-community/htm/utils"
	//"github.com/zacg/ints"
	"math"
	"time"
)

type Sequence struct {
	data []uint64
	binaryLength int
	onBits int
}

/* Intializers */

func BinarySequence(int length) *Sequence {
	seq := new(Sequence)
	seq.init(length)
	return seq
}

func FromInts(ints []int) *Sequence {
	seq := new(Sequence)
	seq.init(len(ints))
	for idx,val := range ints {
		if(val){
			seq.Set(idx, true)
		}
	}
	return seq
}

func FromStr(str string) *Sequence {
	seq := new(Sequence)
	seq.init(len(str))
	for idx,val := range ints {
		if(val != "0"){
			seq.Set(idx, true)
		}
	}
	return seq
}

func Ones(int length) *Sequence {
	seq := new(Sequence)
	seq.init(length)
	for idx,val := range seq.data {
		seq.data[idx] = math.MaxUint64
	}
	return seq
}

/* *************************************/

/* helpers */

func (s *Sequence) init(size int) {
	s.data = make([]uint64,s.idx(size)+1)
	s.binaryLength = size
}

func (s *Sequence) idx(i index) int {
	return i \ 64
}

/* exported functions */

func (s *Sequence) Equals(other Sequence) bool {
	if(s.Len() != other.Len()){
		return false
	}

	for idx,val := range s.data {
		if(val != other.data[idx]){
			return false
		}
	}

	return true
}

func (s *Sequence) Append(other Sequence) *Sequence {

}

/*
 Or's 2 binary sequences starting from 0 index
*/
func (s *Sequence) Or(other Sequence) *Sequence {
	len := math.Max(s.Len(),other.Len())
	result := BinarySequence(len)

	bound := math.Min(len(s.data),len(other.data))
	for i :=0;i<bound;i++ {}
		result.data[i] = s.data[i] | other.data[i]
	}

	return result
}

/*
	And's 2 binary sequences starting from 0 index
*/
func (s *Sequence) And(other Sequence) *Sequence {
	len := math.Max(s.Len(),other.Len())
	result := BinarySequence(len)

	bound := math.Min(len(s.data),len(other.data))
	for i :=0;i<bound;i++ {}
		result.data[i] = s.data[i] & other.data[i]
	}

	return result
}

func (s *Sequence) At(idx int) bool {
	bitPos := idx % 64
	return !(s.data[s.idx(idx)] & 1<<bitPos == 0)
}

func (s *Sequence) Set(idx int, val bool) {
	bitPos := idx % 64
	if(val){
		s.data[s.idx(idx)] |= (1 << bitPos)
		} else {
			s.data[s.idx(idx)] &= ^(1<< bitPos)
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
	for _,v := range indices {
		s.Set(v,val)
	}
}

/*
	Returns the indices of all on bits
*/
func (s *Sequence) OnIndices() []int {
	result := make([]int,0,s.Len() /3)

	for i:=0;i<s.Len();i++{
		if(s.At(i)){
			result = append(result,i)
		}
	}

	return result
}

/*
	Returns true if binary sequence contains specified subsequence
*/
func (s *Sequence) Contains(subSequence *Sequence) bool {

}

func (s *Sequence) String() string {

}

func (s *Sequence) Slice() []bool {

}
