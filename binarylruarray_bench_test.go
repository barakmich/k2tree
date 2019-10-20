package k2tree

import (
	"fmt"
	"testing"
)

var lruBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return newBinaryLRUIndex(&sliceArray{}, 128)
		},
		name: "LRU128",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(&sliceArray{}, 128)
		},
		name: "LRU512",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(128*1024), 128)
		},
		name: "LRU128Paged128kb",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(1024*1024*8), 128)
		},
		name: "LRU128Paged1MB",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(1024*1024*8), 512)
		},
		name: "LRU512Paged1MB",
	},
}

func BenchmarkDetermineLRUCacheDistance(b *testing.B) {
	b.Skip("To determine a good value for DefaultLRUCacheDistance, run this function")
	for _, k2config := range testK2Configs {
		for _, bitarrayt := range lruBitArrayTypes {
			// i is the power of 2 of the size of the cache distance
			for i := 6; i < 13; i++ {
				b.Run(fmt.Sprint(k2config.name, bitarrayt.name, "-", 1<<i), func(b *testing.B) {
					k2, err := newK2Tree(bitarrayt.create, k2config.config)
					if err != nil {
						b.Fatal(err)
					}
					k2.tbits.(*binaryLRUIndex).cacheDistance = 1 << i
					b.ResetTimer()
					populateIncrementalTree(b.N, k2)
				})
			}
		}
	}
}
