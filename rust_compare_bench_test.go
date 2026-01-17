package k2tree

import (
	"testing"
)

// Benchmarks to match the Rust benchmarks exactly

func BenchmarkRandPop1kSlice(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return &sliceArray{} }, SixteenSixteenConfig)
		populateRandomTree(1000, 2000, k2, false)
	}
}

func BenchmarkIncPop1kSlice(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return &sliceArray{} }, SixteenFourConfig)
		populateIncrementalTree(1000, k2, false)
	}
}

func BenchmarkRandPop10kSlice(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return &sliceArray{} }, SixteenSixteenConfig)
		populateRandomTree(10000, 20000, k2, false)
	}
}

func BenchmarkIncPop10kSlice(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return &sliceArray{} }, SixteenFourConfig)
		populateIncrementalTree(10000, k2, false)
	}
}
