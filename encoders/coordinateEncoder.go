package encoders

import (
	"fmt"
	"github.com/zacg/floats"
	"github.com/zacg/htm/utils"
	"math/rand"
	"strconv"
	"strings"
)

/*
 Given a coordinate in an N-dimensional space, and a radius around
that coordinate, the Coordinate Encoder returns an SDR representation
of that position.

The Coordinate Encoder uses an N-dimensional integer coordinate space.
For example, a valid coordinate in this space is (150, -49, 58), whereas
an invalid coordinate would be (55.4, -5, 85.8475).

It uses the following algorithm:

1. Find all the coordinates around the input coordinate, within the
specified radius.
2. For each coordinate, use a uniform hash function to
deterministically map it to a real number between 0 and 1. This is the
"order" of the coordinate.
3. Of these coordinates, pick the top W by order, where W is the
number of active bits desired in the SDR.
4. For each of these W coordinates, use a uniform hash function to
deterministically map it to one of the bits in the SDR. Make this bit active.
5. This results in a final SDR with exactly W bits active
(barring chance hash collisions).
*/
type CoordinateEncoder struct {
	//Number of active bits in SDR
	ActiveBits int
	Width      int
}

func NewCoordinateEncoder(activeBits int, width int) *CoordinateEncoder {
	if activeBits <= 0 || activeBits%2 == 0 {
		panic("param activebits must be an odd positive integer.")
	}

	if width <= 6*activeBits {
		panic("param width must be at least 6 times width. Ideally 11 times greater.")
	}

	c := new(CoordinateEncoder)
	c.ActiveBits = activeBits
	c.Width = width
	return c
}

func (e *CoordinateEncoder) GetName() string {
	return fmt.Sprintf("[%v:%v]", e.Width, e.ActiveBits)
}

// func (e *CoordinateEncoder) GetDescription() string {
// 	return fmt.Sprintf("", )
// }

//convert coord to hashable value
func coordInt(coord []int) int {
	v := 1
	for _, val := range coord {
		v *= val
	}
	return v
}

//returns a coords order
func order(coord []int) float64 {
	v := coordInt(coord)
	//TODO: use different hash function
	rand.Seed(int64(v))
	return rand.Float64()
}

//Map coordinate to active bit index
func (e *CoordinateEncoder) coordBit(coord []int) int {
	v := coordInt(coord)
	rand.Seed(int64(v))
	return rand.Intn(e.Width)
}

func (e *CoordinateEncoder) Encode(input []string, output []bool) {
	if len(input) != 2 {
		panic("invalid coordinate encoder input")
	}

	//parse coordinates
	cordstr := strings.Split(input[0], ",")
	coords := make([]int, 0, len(cordstr))

	for _, val := range cordstr {
		i, err := strconv.Atoi(val)
		if err != nil {
			panic("invalid coordinate")
		}
		coords = append(coords, i)
	}

	radf, err := strconv.ParseFloat(input[1], 64)
	if err != nil {
		panic("invalid radius input for coordinate encoder.")
	}
	radius := int(utils.RoundPrec(radf, 0))

	ranges := make([][]int, len(coords))
	//calc neighbors
	for idx, val := range coords {
		for i := val - radius; i < val+radius+1; i++ {
			ranges[idx] = append(ranges[idx], i)
		}
	}

	neighbors := utils.CartProductInt(ranges)
	//select random top w neighbors
	orders := make([]float64, len(neighbors))
	for _, neighbor := range neighbors {
		orders = append(orders, order(neighbor))
	}
	//sort by order
	indices := make([]int, len(orders))
	floats.Argsort(orders, indices)

	//reset output
	utils.FillSliceBool(output, false)

	//winners
	winners := make([][]int, e.ActiveBits)
	for i := 0; i < e.ActiveBits; i++ {
		winners[i] = neighbors[indices[i]]
	}

	//select top n winners and project bit positions on to result
	for _, val := range winners {
		output[e.coordBit(val)] = true
	}

}
