//
// Code related to temporal pooler stats
//

package htm

import (
	//"fmt"
	//"github.com/cznic/mathutil"
	"github.com/zacg/floats"
	//"github.com/zacg/go.matrix"
	//"math"
	//"math/rand"
	//"sort"
)

type TpStats struct {
	NInfersSinceReset       int
	NPredictions            int
	PredictionScoreTotal    float64
	PredictionScoreTotal2   float64
	FalseNegativeScoreTotal float64
	FalsePositiveScoreTotal float64
	PctExtraTotal           float64
	PctMissingTotal         float64
	TotalMissing            float64
	TotalExtra              float64

	CurPredictionScore    float64
	CurPredictionScore2   float64
	CurFalseNegativeScore float64
	CurFalsePositiveScore float64
	CurMissing            float64
	CurExtra              float64
}

/*
 This function produces goodness-of-match scores for a set of input patterns,
by checking for their presence in the current and predicted output of the
TP. Returns a global count of the number of extra and missing bits, the
confidence scores for each input pattern, and (if requested) the
bits in each input pattern that were not present in the TP's prediction.

@param patternNZs a list of input patterns that we want to check for. Each
element is a list of the non-zeros in that pattern.
@param output The output of the TP. If not specified, then use the
TP's current output. This can be specified if you are
trying to check the prediction metric for an output from
the past.
@param colConfidence The column confidences. If not specified, then use the
TP's current self.colConfidence. This can be specified if you
are trying to check the prediction metrics for an output
from the past.
@param details if True, also include details of missing bits per pattern.

@returns list containing:

[
totalExtras,
totalMissing,
[conf_1, conf_2, ...],
[missing1, missing2, ...]
]

@retval totalExtras a global count of the number of 'extras', i.e. bits that
are on in the current output but not in the or of all the
passed in patterns
@retval totalMissing a global count of all the missing bits, i.e. the bits
that are on in the or of the patterns, but not in the
current output
@retval conf_i the confidence score for the i'th pattern inpatternsToCheck
This consists of 3 items as a tuple:
(predictionScore, posPredictionScore, negPredictionScore)
@retval missing_i the bits in the i'th pattern that were missing
in the output. This list is only returned if details is
True.
*/

type confidence struct {
	PredictionScore         float64
	PositivePredictionScore float64
	NegativePredictionScore float64
}

func (tp *TemporalPooler) checkPrediction2(patternNZs [][]int, output *SparseBinaryMatrix,
	colConfidence []float64, details bool) (int, int, []confidence, []int) {

	// Get the non-zeros in each pattern
	numPatterns := len(patternNZs)

	// Compute the union of all the expected patterns
	var orAll []int
	for _, row := range patternNZs {
		for _, col := range row {
			if !ContainsInt(col, orAll) {
				orAll = append(orAll, col)
			}
		}
	}

	var outputIdxs []int

	// Get the list of active columns in the output
	if output == nil {
		if tp.CurrentOutput == nil {
			panic("Expected tp output")
		}
		outputIdxs = tp.CurrentOutput.NonZeroRows()
	} else {
		outputIdxs = output.NonZeroRows()
	}

	// Compute the total extra and missing in the output
	totalExtras := 0
	totalMissing := 0

	for _, val := range outputIdxs {
		if !ContainsInt(val, orAll) {
			totalExtras++
		}
	}

	for _, val := range orAll {
		if !ContainsInt(val, outputIdxs) {
			totalMissing++
		}
	}

	// Get the percent confidence level per column by summing the confidence
	// levels of the cells in the column. During training, each segment's
	// confidence number is computed as a running average of how often it
	// correctly predicted bottom-up activity on that column. A cell's
	// confidence number is taken from the first active segment found in the
	// cell. Note that confidence will only be non-zero for predicted columns.

	if colConfidence == nil {
		copy(colConfidence, tp.DynamicState.colConfidence)
	}

	// Assign confidences to each pattern
	var confidences []confidence

	for i := 0; i < numPatterns; i++ {
		// Sum of the column confidences for this pattern
		//positivePredictionSum = colConfidence[patternNZs[i]].sum()
		positivePredictionSum := floats.Sum(floats.SubSet(colConfidence, patternNZs[i]))

		// How many columns in this pattern
		positiveColumnCount := len(patternNZs[i])

		// Sum of all the column confidences
		totalPredictionSum := floats.Sum(colConfidence)
		// Total number of columns
		totalColumnCount := len(colConfidence)

		negativePredictionSum := totalPredictionSum - positivePredictionSum
		negativeColumnCount := totalColumnCount - positiveColumnCount

		positivePredictionScore := 0.0
		// Compute the average confidence score per column for this pattern
		if positiveColumnCount != 0 {
			positivePredictionScore = positivePredictionSum
		}

		// Compute the average confidence score per column for the other patterns
		negativePredictionScore := 0.0
		if negativeColumnCount != 0 {
			negativePredictionScore = negativePredictionSum
		}

		// Scale the positive and negative prediction scores so that they sum to
		// 1.0
		currentSum := negativePredictionScore + positivePredictionScore
		if currentSum > 0 {
			positivePredictionScore *= 1.0 / currentSum
			negativePredictionScore *= 1.0 / currentSum
		}

		predictionScore := positivePredictionScore - negativePredictionScore
		newConf := confidence{predictionScore, positivePredictionScore, negativePredictionScore}
		confidences = append(confidences, newConf)

	}

	// Include detail? (bits in each pattern that were missing from the output)
	if details {
		var missingPatternBits []int
		for _, pattern := range patternNZs {
			for _, val := range pattern {
				if !ContainsInt(val, outputIdxs) &&
					!ContainsInt(val, missingPatternBits) {
					missingPatternBits = append(missingPatternBits, val)
				}
			}

		}
		return totalExtras, totalMissing, confidences, missingPatternBits
	} else {
		return totalExtras, totalMissing, confidences, nil
	}

}
