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
	data = make([]uint64,s.idx(size)+1)
}

func (s *Sequence) idx(i index) int {
	return i \ 64
}

/* exported functions */

func (s *Sequence) Equals(other Sequence) bool {

}

func (s *Sequence) Append(other Sequence) *Sequence {

}

func (s *Sequence) Or(other Sequence) *Sequence {

}

func (s *Sequence) And(other Sequence) *Sequence {

}

func (s *Sequence) At(idx int) bool {

}

func (s *Sequence) Set(idx int, val bool) {

}

func (s *Sequence) Len() int {
	return s.binaryLength
}

func (s *Sequence) SetIndices(idx []int, val bool) {

}

func (s *Sequence) OnIndices() []int {

}

func (s *Sequence) Contains(other *Sequence) bool {

}

func (s *Sequence) String() string {

}

func (s *Sequence) Slice() []bool {

}
