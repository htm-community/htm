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
	//"math/rand"
	//"strconv"
	"testing"
)

/*
 Test boosting.
The test is constructed as follows: we construct a set of 5 known inputs. Two
of the input patterns have 50% overlap while all other combinations have 0%
overlap. Each input pattern has 20 bits on to ensure reasonable overlap with
almost all columns.

SP parameters: the minActiveDutyCycle is set to 1 in 10. This allows us to
test boosting with a small number of iterations. The SP is set to have 600
columns with 10% output sparsity. This ensures that the 5 inputs cannot use up
all the columns. Yet we still can have a reasonable number of winning columns
at each step in order to test overlap properties. maxBoost is set to 10 so
that some boosted columns are guaranteed to win eventually but not necessarily
quickly. potentialPct is set to 0.9 to ensure all columns have at least some
overlap with at least one input bit. Thus, when sufficiently boosted, every
column should become a winner at some point. We set permanence increment
and decrement to 0 so that winning columns don't change unless they have
been boosted.

Phase 1: As learning progresses through the first 5 iterations, the first 5
patterns should get distinct output SDRs. The two overlapping input patterns
should have reasonably overlapping output SDRs. The other pattern
combinations should have very little overlap. The boost factor for all
columns should be at 1. At this point least half of the columns should have
never become active and these columns should have duty cycle of 0. Any
columns which have won, should have duty cycles >= 0.2.

Phase 2: Over the next 45 iterations, boosting should stay at 1 for all
columns since minActiveDutyCycle is only calculated after 50 iterations. The
winning columns should be very similar (identical?) to the columns that won
before. About half of the columns should never become active. At the end of
the this phase, most of these columns should have activity level around 0.2.
It's ok for some columns to have higher activity levels.

Phase 3: At this point about half or fewer columns have never won. These
should get boosted to maxBoost and start to win. As each one wins, their
boost gets lowered to 1. After 2 batches, the number of columns that
have never won should be 0. Because of the artificially induced thrashing
behavior in this test, all the inputs should now have pretty distinct
patterns. During this process, as soon as a new column wins, the boost value
for that column should be set back to 1.

Phase 4: Run for 5 iterations without learning on. Boost values and winners
should not change.
*/

type boostTest struct {
	sp               *SpatialPooler
	x                [][]bool
	winningIteration []int
	lastSDR          [][]bool
}

//Returns overlap of the 2 specified sdrs
func computeOverlap(x, y []bool) int {
	result := 0
	for idx, val := range x {
		if val && y[idx] {
			result++
		}
	}
	return result
}

func verifySDRProps(t *testing.T, bt *boostTest) {
	/*
		 Verify that all SDRs have the properties desired for this test.
		The bounds for checking overlap are set fairly loosely here since there is
		some variance due to randomness and the artificial parameters used in this
		test.
	*/

	// Verify that all SDR's are unique
	for i := 0; i <= 4; i++ {
		for j := 1; j <= 4; j++ {
			eq := 0
			for k := 0; k < len(bt.lastSDR[i]); k++ {
				if bt.lastSDR[i][k] == bt.lastSDR[j][k] {
					eq++
				}
			}
			if eq == len(bt.lastSDR[i]) {
				//equal
				assert.Fail(t, "All SDR's are not unique")
			}
		}
	}

	//Verify that the first two SDR's have some overlap.
	expected := computeOverlap(bt.lastSDR[0], bt.lastSDR[1]) > 9
	assert.True(t, expected, "First two SDR's don't overlap much")

	// Verify the last three SDR's have low overlap with everyone else.
	for i := 2; i <= 4; i++ {
		for j := 0; j <= 4; j++ {
			if i != j {
				overlap := computeOverlap(bt.lastSDR[i], bt.lastSDR[j])
				expected := overlap < 18
				assert.True(t, expected, "One of the last three SDRs has high overlap")
			}
		}
	}

}

func phase1(t *testing.T, bt *boostTest) {
	y := make([]bool, bt.sp.numColumns)
	// Do one training batch through the input patterns
	for idx, input := range bt.x {
		utils.FillSliceBool(y, false)
		bt.sp.Compute(input, true, y, bt.sp.InhibitColumns)
		for j, winner := range y {
			if winner {
				bt.winningIteration[j] = bt.sp.IterationLearnNum
			}
		}
		bt.lastSDR[idx] = y
	}

	//The boost factor for all columns should be at 1.
	assert.Equal(t, bt.sp.numColumns, utils.CountFloat64(bt.sp.boostFactors, 1), "Boost factors are not all 1")

	//At least half of the columns should have never been active.
	winners := utils.CountInt(bt.winningIteration, 0)
	assert.True(t, winners >= bt.sp.numColumns/2, "More than half of the columns have been active")

	//All the never-active columns should have duty cycle of 0
	//All the at-least-once-active columns should have duty cycle >= 0.2
	activeSum := 0.0
	for idx, val := range bt.sp.activeDutyCycles {
		if bt.winningIteration[idx] == 0 {
			//assert.Equal(t, expected, actual, ...)
			activeSum += val
		}
	}
	assert.Equal(t, 0, activeSum, "Inactive columns have positive duty cycle.")

	winningMin := 100000.0
	for idx, val := range bt.sp.activeDutyCycles {
		if bt.winningIteration[idx] > 0 {
			if val < winningMin {
				winningMin = val
			}
		}
	}
	assert.True(t, winningMin >= 0.2, "Active columns have duty cycle that is too low.")

	verifySDRProps(t, bt)
}

func phase2(t *testing.T, bt *boostTest) {

	y := make([]bool, bt.sp.numColumns)

	// Do 9 training batch through the input patterns
	for i := 0; i < 9; i++ {
		for idx, input := range bt.x {
			utils.FillSliceBool(y, false)
			bt.sp.Compute(input, true, y, bt.sp.InhibitColumns)
			for j, winner := range y {
				if winner {
					bt.winningIteration[j] = bt.sp.IterationLearnNum
				}
			}
			bt.lastSDR[idx] = y
		}
	}

	// The boost factor for all columns should be at 1.
	assert.Equal(t, bt.sp.numColumns, utils.CountFloat64(bt.sp.boostFactors, 1), "Boost factors are not all 1")

	// Roughly half of the columns should have never been active.
	winners := utils.CountInt(bt.winningIteration, 0)
	assert.True(t, winners >= int(0.4*float64(bt.sp.numColumns)), "More than 60% of the columns have been active")

	// All the never-active columns should have duty cycle of 0
	activeSum := 0.0
	for idx, val := range bt.sp.activeDutyCycles {
		if bt.winningIteration[idx] == 0 {
			activeSum += val
		}
	}
	assert.Equal(t, 0, activeSum, "Inactive columns have positive duty cycle.")

	dutyAvg := 0.0
	dutyCount := 0
	for _, val := range bt.sp.activeDutyCycles {
		if val > 0 {
			dutyAvg += val
			dutyCount++
		}
	}

	// The average at-least-once-active columns should have duty cycle >= 0.15
	// and <= 0.25
	dutyAvg = dutyAvg / float64(dutyCount)
	assert.True(t, dutyAvg >= 0.15, "Average on-columns duty cycle is too low.")
	assert.True(t, dutyAvg <= 0.30, "Average on-columns duty cycle is too high.")

	verifySDRProps(t, bt)
}

func phase3(t *testing.T, bt *boostTest) {
	//Do two more training batches through the input patterns
	y := make([]bool, bt.sp.numColumns)

	for i := 0; i < 2; i++ {
		for idx, input := range bt.x {
			utils.FillSliceBool(y, false)
			bt.sp.Compute(input, true, y, bt.sp.InhibitColumns)
			for j, winner := range y {
				if winner {
					bt.winningIteration[j] = bt.sp.IterationLearnNum
				}
			}
			bt.lastSDR[idx] = y
		}
	}

	// The boost factor for all columns that just won should be at 1.
	for idx, val := range y {
		if val {
			if bt.sp.boostFactors[idx] != 1 {
				assert.Fail(t, "Boost factors of winning columns not 1")
			}
		}
	}

	// By now, every column should have been sufficiently boosted to win at least
	// once. The number of columns that have never won should now be 0
	for _, val := range bt.winningIteration {
		if val == 0 {
			assert.Fail(t, "Expected all columns to have won atleast once.")
		}
	}

	// Because of the artificially induced thrashing, even the first two patterns
	// should have low overlap. Verify that the first two SDR's now have little
	// overlap
	overlap := computeOverlap(bt.lastSDR[0], bt.lastSDR[1])
	assert.True(t, overlap < 7, "First two SDR's overlap significantly when they should not")
}

func phase4(t *testing.T, bt *boostTest) {
	//The boost factor for all columns that just won should be at 1.
	boostAtBeg := make([]float64, len(bt.sp.boostFactors))
	copy(bt.sp.boostFactors, boostAtBeg)

	// Do one more iteration through the input patterns with learning OFF
	y := make([]bool, bt.sp.numColumns)
	for _, input := range bt.x {
		utils.FillSliceBool(y, false)
		bt.sp.Compute(input, false, y, bt.sp.InhibitColumns)

		// The boost factor for all columns that just won should be at 1.
		assert.Equal(t, utils.SumSliceFloat64(boostAtBeg), utils.SumSliceFloat64(bt.sp.boostFactors), "Boost factors changed when learning is off")
	}

}

func BoostTest(t *testing.T) {
	bt := boostTest{}
	spParams := NewSpParams()
	spParams.InputDimensions = []int{90}
	spParams.ColumnDimensions = []int{600}
	spParams.PotentialRadius = 90
	spParams.PotentialPct = 0.9
	spParams.GlobalInhibition = true
	spParams.NumActiveColumnsPerInhArea = 60
	spParams.MinPctActiveDutyCycle = 0.1
	spParams.SynPermActiveInc = 0
	spParams.SynPermInactiveDec = 0
	spParams.DutyCyclePeriod = 10
	bt.sp = NewSpatialPooler(spParams)

	// Create a set of input vectors, x
	// B,C,D don't overlap at all with other patterns
	bt.x = make([][]bool, 5)
	for i := range bt.x {
		bt.x[i] = make([]bool, bt.sp.numInputs)
	}

	utils.FillSliceRangeBool(bt.x[0], true, 0, 20)
	utils.FillSliceRangeBool(bt.x[1], true, 10, 30)
	utils.FillSliceRangeBool(bt.x[2], true, 30, 50)
	utils.FillSliceRangeBool(bt.x[3], true, 50, 70)
	utils.FillSliceRangeBool(bt.x[4], true, 70, 90)
	// For each column, this will contain the last iteration number where that
	// column was a winner
	bt.winningIteration = make([]int, bt.sp.numColumns)

	// For each input vector i, lastSDR[i] contains the most recent SDR output
	// by the SP.
	bt.lastSDR = make([][]bool, 5)

	phase1(t, &bt)
	phase2(t, &bt)
	phase3(t, &bt)
	phase4(t, &bt)
}
