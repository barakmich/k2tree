package k2tree

import (
	"testing"
)

// xorshift32 PRNG — identical algorithm to the Rust version so both benchmarks
// use the same sequence of pseudo-random numbers from the same seed.
type benchPRNG struct {
	state uint32
}

func (p *benchPRNG) next() uint32 {
	p.state ^= p.state << 13
	p.state ^= p.state >> 17
	p.state ^= p.state << 5
	return p.state
}

func (p *benchPRNG) nextN(n uint32) uint32 {
	return p.next() % n
}

// populate50kSeeded fills k2 with 50 000 random links drawn from the same
// seed every time, so the tree is identical across benchmark runs and across
// the Go/Rust comparison.
func populate50kSeeded(k2 *K2Tree) {
	rng := benchPRNG{state: 12345}
	for i := 0; i < 50000; i++ {
		row := int(rng.nextN(100000))
		col := int(rng.nextN(100000))
		k2.Add(row, col)
	}
}

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

// QuartileIndex benchmarks to match Rust

func BenchmarkRandPop1kQuartile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenSixteenConfig)
		populateRandomTree(1000, 2000, k2, false)
	}
}

func BenchmarkIncPop1kQuartile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenFourConfig)
		populateIncrementalTree(1000, k2, false)
	}
}

func BenchmarkRandPop10kQuartile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenSixteenConfig)
		populateRandomTree(10000, 20000, k2, false)
	}
}

func BenchmarkIncPop10kQuartile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenFourConfig)
		populateIncrementalTree(10000, k2, false)
	}
}

func BenchmarkRandPop50kQuartile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenSixteenConfig)
		populateRandomTree(50000, 100000, k2, false)
	}
}

func BenchmarkIncPop50kQuartile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenFourConfig)
		populateIncrementalTree(50000, k2, false)
	}
}

// Iteration benchmarks: populate once with a fixed seed, then measure how
// fast we can walk every row's forward links across the full node space.
// (Reverse / column iteration is not yet implemented in either language.)

func BenchmarkIterateAll50kSlice(b *testing.B) {
	k2, _ := newK2Tree(func() bitarray { return &sliceArray{} }, SixteenSixteenConfig)
	populate50kSeeded(k2)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var sum int
		for row := 0; row < 100000; row++ {
			it := newRowIterator(k2, row)
			for it.Next() {
				sum++
			}
		}
		_ = sum
	}
}

func BenchmarkIterateAll50kQuartile(b *testing.B) {
	k2, _ := newK2Tree(func() bitarray { return newQuartileIndex(&sliceArray{}) }, SixteenSixteenConfig)
	populate50kSeeded(k2)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var sum int
		for row := 0; row < 100000; row++ {
			it := newRowIterator(k2, row)
			for it.Next() {
				sum++
			}
		}
		_ = sum
	}
}
