htm
===

Hierarchical Temporal Memory Implementation in Golang

This is a direct port of the spatial and temporal poolers as they currently exist in Numentas Nupic Project. The Nupic project basically demonstrates a single stage of the cortical hierarchy. Eventually this same code can be extended to form a full HTM hierarchy.

##Changes
 * Temporal pooler ephemeral state is stored in strongly typed struct rather than a hashmap. t-1 vars have "last" appended to their names.
 * Temporal pooler params stored in "params" sub struct
 * Binary data structures are used rather than ints
 * No C++ dependency everything is written in Go

##Current State of Project
 * Spatial pooler mostly complete, all unit tests passing
 * Temporal pooler under construction

##Outstanding
 * Finish temporal pooler and related unit tests
 * Implement a better sparse binary matrix structure with versions optimized for col or row heavy access.
 * Refactor to be more idiomatic Go. It is currently basically a line for line port of the python implementation, it could be refactored to make better use of Go's type system.
