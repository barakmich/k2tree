package k2tree

import "testing"

func BenchmarkIterateAll1kSlice(b *testing.B) {
	k2, _ := newK2Tree(func() bitarray { return &sliceArray{} }, SixteenSixteenConfig)
	populateRandomTree(1000, 2000, k2, false)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var sum int
		for row := 0; row < k2.maxIndex(); row++ {
			it := newRowIterator(k2, row)
			for it.Next() {
				sum++
			}
		}
		_ = sum
	}
}
