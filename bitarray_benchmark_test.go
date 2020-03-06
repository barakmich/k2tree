package k2tree

import (
	"fmt"
	"testing"

	"git.barakmich.com/barak/k2tree/bytearray"
)

func testPopulateRand(t testing.TB, ba newBitArrayFunc, n int, compare bool) *K2Tree {
	createF := ba
	if compare {
		createF = func() bitarray {
			x := ba()
			return newCompareArray(&sliceArray{}, x)
		}
	}
	k2, err := newK2Tree(
		createF,
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: SixteenBitsPerLayer,
		})
	if err != nil {
		t.Fatal(err)
	}
	populateRandomTree(n, n*2, k2)
	return k2
}

func testPopulateIncremental(t testing.TB, ba newBitArrayFunc, n int, compare bool) *K2Tree {
	createF := ba
	if compare {
		createF = func() bitarray {
			x := ba()
			return newCompareArray(&sliceArray{}, x)
		}
	}
	k2, err := newK2Tree(
		createF,
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: SixteenBitsPerLayer,
		})
	if err != nil {
		t.Fatal(err)
	}
	populateIncrementalTree(n, k2)
	return k2
}

func TestRandPop50k(t *testing.T) {
	for _, bitarray := range testBitArrayTypes {
		t.Run(fmt.Sprintf(bitarray.name), func(t *testing.T) {
			k2 := testPopulateRand(t, bitarray.create, 50000, true)
			t.Logf("%f bpl", k2.Stats().BitsPerLink)
		})
	}
}

func TestIncPop50k(t *testing.T) {
	for _, bitarray := range testBitArrayTypes {
		t.Run(fmt.Sprintf(bitarray.name), func(t *testing.T) {
			k2 := testPopulateIncremental(t, bitarray.create, 50000, true)
			t.Logf("%f bpl", k2.Stats().BitsPerLink)
		})
	}
}

func BenchmarkRandPop50k(b *testing.B) {
	for _, bitarray := range testBitArrayTypes {
		b.Run(fmt.Sprintf(bitarray.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				testPopulateRand(b, bitarray.create, 50000, true)
			}
		})
	}
}

func BenchmarkRandPop100k(b *testing.B) {
	for _, bitarray := range fastBitArrayTypes {
		b.Run(fmt.Sprintf(bitarray.name), func(b *testing.B) {
			var k2 *K2Tree
			for n := 0; n < b.N; n++ {
				k2 = testPopulateRand(b, bitarray.create, 100000, false)
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			}
		})
	}
}

func BenchmarkIncPop1M(b *testing.B) {
	for _, bitarrayt := range fastBitArrayTypes {
		b.Run(fmt.Sprintf(bitarrayt.name), func(b *testing.B) {
			var k2 *K2Tree
			for n := 0; n < b.N; n++ {
				var err error
				k2, err = newK2Tree(
					bitarrayt.create,
					Config{
						TreeLayerDef: SixteenBitsPerLayer,
						CellLayerDef: SixteenBitsPerLayer,
					})
				if err != nil {
					b.Fatal(err)
				}
				populateIncrementalTree(1000000, k2)
			}
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
		})
	}
}

func BenchmarkIncPopVar(b *testing.B) {
	for _, k2config := range testK2Configs {
		for _, bitarrayt := range fastBitArrayTypes {
			b.Run(fmt.Sprint(k2config.name, bitarrayt.name), func(b *testing.B) {
				k2, err := newK2Tree(bitarrayt.create, k2config.config)
				if err != nil {
					b.Fatal(err)
				}
				b.ResetTimer()
				populateIncrementalTree(b.N, k2)
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			})
		}

	}
}

func BenchmarkRandPopVar(b *testing.B) {
	for _, k2config := range testK2Configs {
		for _, bitarrayt := range fastBitArrayTypes {
			b.Run(fmt.Sprint(k2config.name, bitarrayt.name), func(b *testing.B) {
				k2, err := newK2Tree(bitarrayt.create, k2config.config)
				if err != nil {
					b.Fatal(err)
				}
				b.ResetTimer()
				populateRandomTree(b.N, b.N*2, k2)
				stats := k2.Stats()
				b.ReportMetric(stats.BitsPerLink, "bits/link")
			})
		}

	}
}

func BenchmarkRandUnindexed(b *testing.B) {
	for _, bitarrayt := range unindexedBitArrayTypes {
		b.Run(bitarrayt.name, func(b *testing.B) {
			k2, err := newK2Tree(bitarrayt.create, SixteenSixteenConfig)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			populateRandomTree(b.N, b.N*2, k2)
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
		})
	}
}

func BenchmarkIncUnindexed(b *testing.B) {
	for _, bitarrayt := range unindexedBitArrayTypes {
		b.Run(bitarrayt.name, func(b *testing.B) {
			k2, err := newK2Tree(bitarrayt.create, SixteenSixteenConfig)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			populateIncrementalTree(b.N, k2)
			stats := k2.Stats()
			b.ReportMetric(stats.BitsPerLink, "bits/link")
		})
	}
}

var unindexedBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return &sliceArray{}
		},
		name: "SliceArray",
	},
	{
		create: func() bitarray {
			return newPagedSliceArray(128 * 1024)
		},
		name: "Paged128kb",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewSlice())
		},
		name: "ByteArray:Slice",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewFrontSlice(1024))
		},
		name: "ByteArray:Front",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewSpillover(4096, 0.8, 0.3, false))
		},
		name: "ByteArray:Spill:4096:80:30:1x",
	},
	{
		create: func() bitarray {
			return newByteArray(bytearray.NewSpillover(32*1024, 0.8, 0.3, false))
		},
		name: "ByteArray:Spill:32k:80:30:1x",
	},
}

var fastBitArrayTypes []bitArrayType = []bitArrayType{
	{
		create: func() bitarray {
			return newInt16Index(&sliceArray{})
		},
		name: "Int16",
	},
	{
		create: func() bitarray {
			return newInt16Index(newPagedSliceArray(128 * 1024))
		},
		name: "Int16Paged128kb",
	},
	{
		create: func() bitarray {
			return newInt16Index(newByteArray(bytearray.NewSlice()))
		},
		name: "Int16BASlice",
	},
	{
		create: func() bitarray {
			return newInt16Index(newByteArray(bytearray.NewSpillover(4096, 0.8, 0.3, false)))
		},
		name: "Int16BASpill4k1x",
	},
	{
		create: func() bitarray {
			return newInt16Index(newByteArray(bytearray.NewSpillover(32*1024, 0.8, 0.3, false)))
		},
		name: "Int16BASpill32k1x",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(&sliceArray{}, 128)
		},
		name: "LRU128",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newPagedSliceArray(128*1024), 128)
		},
		name: "LRU128Paged128kb",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSlice()), 128)
		},
		name: "LRU128BASlice",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSpillover(4096, 0.8, 0.3, false)), 128)
		},
		name: "LRU128BASpill4k1x",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSpillover(32*1024, 0.8, 0.3, false)), 128)
		},
		name: "LRU128BASpill32k1x",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSpillover(4096, 0.9, 0.75, false)), 128)
		},
		name: "LRU128BASpill4k1xh",
	},
	{
		create: func() bitarray {
			return newBinaryLRUIndex(newByteArray(bytearray.NewSpillover(32*1024, 0.9, 0.75, false)), 128)
		},
		name: "LRU128BASpill32k1xh",
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

type k2configTest struct {
	config Config
	name   string
}

var testK2Configs []k2configTest = []k2configTest{
	{
		config: SixteenFourConfig,
		name:   "16x4",
	},
	{
		config: SixteenSixteenConfig,
		name:   "16x16",
	},
	//{
	//config: SixtySixteenConfig,
	//name:   "64x16",
	//},
	//{
	//config: Config{
	//TreeLayerDef: TwoFiftySixBitsPerLayer,
	//CellLayerDef: SixteenBitsPerLayer,
	//},
	//name: "256x16",
	//},
}
