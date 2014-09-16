htm
===

Hierarchical Temporal Memory Implementation in Golang

[![GoDoc](https://godoc.org/github.com/zacg/htm?status.png)](https://godoc.org/github.com/zacg/htm)
[![Build Status](https://travis-ci.org/zacg/htm.svg?branch=master)](https://travis-ci.org/zacg/htm)

This is a direct port of the spatial and temporal poolers as they currently exist in Numenta's Nupic Project. This project was done as a learning exercise, no effort has been made to optimize this implementation and it was not designed for production use.

The Nupic project basically demonstrates the CLA, a single stage of the cortical hierarchy. Eventually this same code can be extended to form a full HTM hierarchy. https://github.com/numenta/nupic

##Changes From Numentas Implementation
 * Temporal pooler ephemeral state is stored in strongly typed struct rather than a hashmap. t-1 vars have "last" appended to their names.
 * Temporal pooler params stored in "params" sub struct
 * Binary data structures are used rather than ints
 * No C++ dependency everything is written in Go

##Current State of Project
 * Temporal and Spatial poolers pass basic tests
 * Temporal memory passes basic unit tests

##Todo
 ~~* Finish temporal unit tests~~
 * Implement a better sparse binary matrix structure with versions optimized for col or row heavy access.
 * Refactor to be more idiomatic Go. It is basically a line for line port of the python implementation, it could be refactored to make better use of Go's type system.
 * Implement some of the common encoders

##Examples

###Temporal Pooler
```go
package main

import (
	"fmt"
	"github.com/zacg/htm"
	"github.com/zacg/htm/utils"
)

func main() {
	tps := htm.NewTemporalPoolerParams()
	tps.Verbosity = 0
	tps.NumberOfCols = 50
	tps.CellsPerColumn = 2
	tps.ActivationThreshold = 8
	tps.MinThreshold = 10
	tps.InitialPerm = 0.5
	tps.ConnectedPerm = 0.5
	tps.NewSynapseCount = 10
	tps.PermanenceDec = 0.0
	tps.PermanenceInc = 0.1
	tps.GlobalDecay = 0
	tps.BurnIn = 1
	tps.PamLength = 10
	tps.CollectStats = true
	tp := htm.NewTemporalPooler(*tps)

	//Mock encoding of ABCDE
	inputs := make([][]bool, 5)
	inputs[0] = boolRange(0, 9, 50)   //bits 0-9 are "on"
	inputs[1] = boolRange(10, 19, 50) //bits 10-19 are "on"
	inputs[2] = boolRange(20, 29, 50) //bits 20-29 are "on"
	inputs[3] = boolRange(30, 39, 50) //bits 30-39 are "on"
	inputs[4] = boolRange(40, 49, 50) //bits 40-49 are "on"

	//Learn 5 sequences above
	for i := 0; i < 10; i++ {
		for p := 0; p < 5; p++ {
			tp.Compute(inputs[p], true, false)
		}
		tp.Reset()
	}

	//Predict sequences
	for i := 0; i < 4; i++ {
		tp.Compute(inputs[i], false, true)
		p := tp.DynamicState.InfPredictedState

		fmt.Printf("Predicted: %v From input: %v \n", p.NonZeroRows(), utils.OnIndices(inputs[i]))

	}

}

//helper method for creating boolean sequences
func boolRange(start int, end int, length int) []bool {
	result := make([]bool, length)
	for i := start; i <= end; i++ {
		result[i] = true
	}
	return result
}


```

###Spatial Pooler
```go
package main

import (
	"fmt"
	"github.com/davecheney/profile"
	"github.com/zacg/htm"
	"github.com/zacg/htm/utils"
	"math/rand"
)

func main() {
	
	ssp := htm.NewSpParams()
	ssp.ColumnDimensions = []int{64, 64}
	ssp.InputDimensions = []int{32, 32}
	ssp.PotentialRadius = ssp.NumInputs()
	ssp.NumActiveColumnsPerInhArea = int(0.02 * float64(ssp.NumColumns()))
	ssp.GlobalInhibition = true
	ssp.SynPermActiveInc = 0.01
	ssp.SpVerbosity = 10
	sp := htm.NewSpatialPooler(ssp)
	

	activeArray := make([]bool, sp.NumColumns())
	inputVector := make([]bool, sp.NumInputs())

	for idx, _ := range inputVector {
		inputVector[idx] = rand.Intn(5) >= 2
	}

	sp.Compute(inputVector, true, activeArray, sp.InhibitColumns)

	fmt.Println("Active Indices:", utils.OnIndices(activeArray))

}

```