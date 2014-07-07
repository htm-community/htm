htm
===

Hierarchical Temporal Memory Implementation in Golang

[![GoDoc](https://godoc.org/github.com/zacg/htm?status.png)](https://godoc.org/github.com/zacg/htm)

This is a direct port of the spatial and temporal poolers as they currently exist in Numenta's Nupic Project. This project was done as a learning exercise, no effort has been made to optimize this implementation and it was not designed for production use.

The Nupic project basically demonstrates a single stage of the cortical hierarchy. Eventually this same code can be extended to form a full HTM hierarchy. https://github.com/numenta/nupic

##Changes From Numentas Implementation
 * Temporal pooler ephemeral state is stored in strongly typed struct rather than a hashmap. t-1 vars have "last" appended to their names.
 * Temporal pooler params stored in "params" sub struct
 * Binary data structures are used rather than ints
 * No C++ dependency everything is written in Go

##Current State of Project
 * Temporal and Spatial poolers pass basic tests

##Todo
 * Finish temporal unit tests
 * Implement a better sparse binary matrix structure with versions optimized for col or row heavy access.
 * Refactor to be more idiomatic Go. It is basically a line for line port of the python implementation, it could be refactored to make better use of Go's type system.


##Examples

{% gist d1640e7346e9747562e7 %}