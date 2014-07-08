package htm

import (
	//"fmt"
	//"github.com/skelterjohn/go.matrix"
	//"github.com/stretchr/testify/assert"
	"github.com/zacg/htm/utils"
	"github.com/zacg/testify/assert"
	//"math/big"
	//"github.com/stretchr/testify/mock"
	//"math"
	"math/rand"
	//"strconv"
	"testing"
)

func basicComputeLoop(t *testing.T, spParams SpParams) {
	/*
		 Feed in some vectors and retrieve outputs. Ensure the right number of
		columns win, that we always get binary outputs, and that nothing crashes.
	*/

	sp := NewSpatialPooler(spParams)

	// Create a set of input vectors as well as various numpy vectors we will
	// need to retrieve data from the SP
	numRecords := 100

	inputMatrix := make([][]bool, numRecords)
	for i := range inputMatrix {
		inputMatrix[i] = make([]bool, sp.numInputs)
		for j := range inputMatrix[i] {
			inputMatrix[i][j] = rand.Float64() > 0.8
		}
	}

	// With learning off and no prior training we should get no winners
	y := make([]bool, sp.numColumns)
	for _, input := range inputMatrix {
		utils.FillSliceBool(y, false)
		sp.Compute(input, false, y, sp.InhibitColumns)
		assert.Equal(t, 0, utils.CountTrue(y))
	}

	// With learning on we should get the requested number of winners
	for _, input := range inputMatrix {
		utils.FillSliceBool(y, false)
		sp.Compute(input, true, y, sp.InhibitColumns)
		assert.Equal(t, sp.NumActiveColumnsPerInhArea, utils.CountTrue(y))

	}

	// With learning off and some prior training we should get the requested
	// number of winners
	for _, input := range inputMatrix {
		utils.FillSliceBool(y, false)
		sp.Compute(input, false, y, sp.InhibitColumns)
		assert.Equal(t, sp.NumActiveColumnsPerInhArea, utils.CountTrue(y))
	}

}

func TestBasicCompute1(t *testing.T) {

	spParams := NewSpParams()
	spParams.InputDimensions = []int{30}
	spParams.ColumnDimensions = []int{50}
	spParams.GlobalInhibition = true

	basicComputeLoop(t, spParams)
}

func TestBasicCompute2(t *testing.T) {

	spParams := NewSpParams()
	spParams.InputDimensions = []int{100}
	spParams.ColumnDimensions = []int{100}
	spParams.GlobalInhibition = true
	spParams.SynPermActiveInc = 0
	spParams.SynPermInactiveDec = 0

	basicComputeLoop(t, spParams)

}
